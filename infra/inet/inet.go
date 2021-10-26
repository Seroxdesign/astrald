package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/ip"
	"log"
	"net"
	"strconv"
	"strings"
)

var _ infra.Network = &Inet{}

type Inet struct {
	listenPort     uint16
	extAddrs       []Addr
	separateListen bool
}

func New() *Inet {
	return &Inet{
		listenPort: defaultPort,
		extAddrs:   make([]Addr, 0),
	}
}

func (inet Inet) Name() string {
	return NetworkName
}

func (inet *Inet) AddExternalAddr(s string) error {
	addr, err := Parse(s)
	if err != nil {
		return err
	}

	inet.extAddrs = append(inet.extAddrs, addr)

	return nil
}

func (inet Inet) Addresses() []infra.AddrDesc {
	list := make([]infra.AddrDesc, 0)

	ifaceAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	for _, a := range ifaceAddrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}

		ipv4 := ipnet.IP.To4()
		if ipv4 == nil {
			continue
		}

		if ipv4.IsLoopback() {
			continue
		}

		addr := Addr{ip: ipv4, port: inet.listenPort}

		list = append(list, infra.AddrDesc{
			Addr:   addr,
			Public: !ipv4.IsPrivate(),
		})
	}

	// Add external addresses
	for _, a := range inet.extAddrs {
		list = append(list, infra.AddrDesc{
			Addr:   a,
			Public: true,
		})
	}

	return list
}

func (inet Inet) Unpack(bytes []byte) (infra.Addr, error) {
	return Unpack(bytes)
}

func (inet Inet) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	a, ok := addr.(Addr)
	if !ok {
		return nil, infra.ErrUnsupportedAddress
	}

	return Dial(ctx, a)
}

func (inet Inet) Listen(ctx context.Context) (<-chan infra.Conn, <-chan error) {
	if inet.separateListen {
		return inet.listenSeparately(ctx)
	}
	return inet.listenCombined(ctx)
}

func (inet Inet) listenCombined(ctx context.Context) (<-chan infra.Conn, <-chan error) {
	output, errCh := make(chan infra.Conn), make(chan error, 1)

	go func() {
		defer close(output)
		defer close(errCh)

		hostPort := "0.0.0.0:" + strconv.Itoa(int(inet.listenPort))

		l, err := net.Listen("tcp", hostPort)
		if err != nil {
			errCh <- err
			return
		}

		log.Println("listen tcp", hostPort)

		go func() {
			<-ctx.Done()
			l.Close()
		}()

		for {
			conn, err := l.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					errCh <- err
				}
				return
			}

			output <- newConn(conn, false)
		}
	}()

	return output, errCh
}

func (inet Inet) listenSeparately(ctx context.Context) (<-chan infra.Conn, <-chan error) {
	output := make(chan infra.Conn)

	go func() {
		defer close(output)

		for ifaceName := range ip.Interfaces(ctx) {
			go func(ifaceName string) {
				for conn := range listenInterface(ctx, ifaceName) {
					output <- conn
				}
			}(ifaceName)
		}
	}()

	return output, nil
}

func (inet Inet) Broadcast(ctx context.Context, payload []byte) <-chan error {
	return errChan(infra.ErrUnsupportedOperation)
}

func (inet Inet) Scan(ctx context.Context) (<-chan infra.Broadcast, <-chan error) {
	return nil, errChan(infra.ErrUnsupportedOperation)
}

func errChan(err error) <-chan error {
	ch := make(chan error, 1)
	defer close(ch)
	ch <- err
	return ch
}
