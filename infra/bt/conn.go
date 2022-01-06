package bt

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/infra"
	"golang.org/x/sys/unix"
	"io"
	"sync"
)

var _ infra.Conn = &Conn{}

type Conn struct {
	mu         sync.RWMutex
	nfd        int
	outbount   bool
	localAddr  Addr
	remoteAddr Addr
}

func (conn Conn) Read(p []byte) (n int, err error) {
	fd := unix.PollFd{
		Fd:      int32(conn.nfd),
		Events:  unix.POLLIN | unix.POLLRDHUP,
		Revents: 0,
	}

	for {
		debugln("(bt) POLL...")
		n, err = unix.Poll([]unix.PollFd{fd}, -1)
		debugln("(bt) POLL", n, err)

		// retry only on interrupt
		if err != unix.EINTR {
			break
		}
	}

	if n != 1 {
		return 0, err
	}

	conn.mu.RLock()
	defer conn.mu.RUnlock()

	debugln("(bt) READ...")
	n, err = unix.Read(conn.nfd, p)
	debugln("(bt) READ", n, err)

	if n < 0 {
		return 0, err
	}
	if n == 0 {
		return 0, io.EOF
	}

	return n, err
}

func (conn Conn) Write(p []byte) (n int, err error) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	debugln("(bt) WRITE...")
	n, err = unix.Write(conn.nfd, p)
	debugln("(bt) WRITTEN", n, err)

	if n < 0 {
		n = 0
		err = fmt.Errorf("write error: %w", err)
		conn.Close()
	}
	return
}

func (conn Conn) Close() error {
	return unix.Shutdown(conn.nfd, unix.SHUT_RDWR)
}

func (conn Conn) Outbound() bool {
	return conn.outbount
}

func (conn Conn) LocalAddr() infra.Addr {
	return conn.localAddr
}

func (conn Conn) RemoteAddr() infra.Addr {
	return conn.remoteAddr
}
