package link

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	"sync"
)

type PortBinding struct {
	*net.OutputField
	target   io.WriteCloser
	link     *CoreLink
	port     int
	capacity int
	used     int
	chunks   [][]byte
	ch       chan struct{}
	chunksMu sync.Mutex
	outputMu sync.RWMutex
}

func NewPortBinding(output net.SecureWriteCloser, link *CoreLink, port int) *PortBinding {
	binding := &PortBinding{
		port:     port,
		link:     link,
		capacity: portBufferSize,
		chunks:   make([][]byte, 0),
		ch:       make(chan struct{}, 1),
	}
	binding.OutputField = net.NewOutputField(binding, output)

	go binding.flusher()

	return binding
}

func (b *PortBinding) HandleMux(event mux.Event) {
	switch event := event.(type) {
	case mux.Frame:
		b.handleFrame(event)

	case mux.Unbind:
		b.pushChunk([]byte{})
	}
}

func (b *PortBinding) Link() *CoreLink {
	return b.link
}

func (b *PortBinding) Transport() net.SecureConn {
	return b.link.Transport()
}

func (b *PortBinding) Port() int {
	return b.port
}

func (b *PortBinding) handleFrame(frame mux.Frame) {
	// register link activity
	b.link.Touch()

	// check EOF
	if frame.IsEmpty() {
		frame.Mux.Unbind(frame.Port)
		return
	}

	// add chunk to the buffer
	if err := b.pushChunk(frame.Data); err != nil {
		b.link.CloseWithError(err)
	}
}

func (b *PortBinding) pushChunk(p []byte) error {
	b.chunksMu.Lock()
	defer b.chunksMu.Unlock()

	if len(p) > b.available() {
		return ErrPortBufferOverflow
	}

	b.chunks = append(b.chunks, p)
	b.used += len(p)

	b.signal()

	return nil
}

func (b *PortBinding) popChunk() ([]byte, error) {
	b.chunksMu.Lock()
	defer b.chunksMu.Unlock()

	if len(b.chunks) == 0 {
		return nil, ErrPortBufferEmpty
	}

	chunk := b.chunks[0]
	b.chunks = b.chunks[1:]
	b.used -= len(chunk)

	return chunk, nil
}

func (b *PortBinding) chunkCount() int {
	b.chunksMu.Lock()
	defer b.chunksMu.Unlock()

	return len(b.chunks)
}

func (b *PortBinding) signal() {
	select {
	case b.ch <- struct{}{}:
	default:
	}
}

func (b *PortBinding) flusher() {
	defer func() {
		b.outputMu.Lock()
		defer b.outputMu.Unlock()

		b.OutputField.Output().Close()
		b.link.control.Reset(b.port)
	}()

	for {
		b.wait()
		var flushed int

		for {
			chunk, err := b.popChunk()
			if err != nil {
				break
			}

			// EOF?
			if len(chunk) == 0 {
				return
			}

			n, err := b.write(chunk)
			if len(chunk) != n {
				b.link.CloseWithError(errors.New("partial write on port"))
				return
			}
			if err != nil {
				b.link.CloseWithError(err)
				return
			}

			flushed += n
			if flushed >= defaultMaxFrameSize {
				break
			}
		}

		if flushed > 0 {
			_ = b.link.control.GrowBuffer(b.port, flushed)
		}
	}
}

func (b *PortBinding) write(p []byte) (int, error) {
	b.outputMu.RLock()
	defer b.outputMu.RUnlock()

	return b.OutputField.Output().Write(p)
}

func (b *PortBinding) wait() error {
	if b.chunkCount() > 0 {
		return nil
	}
	<-b.ch
	return nil
}

func (b *PortBinding) available() int {
	return b.capacity - b.used
}
