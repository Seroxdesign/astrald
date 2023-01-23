package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/config"
	"io"
	"strings"
	"sync"
)

type Querier interface {
	Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error)
}

type Infra struct {
	Querier  Querier
	networks map[string]infra.Network
	localID  id.Identity
	rootDir  string

	gateways  []infra.AddrSpec
	inet      *inet.Inet
	tor       *tor.Tor
	gateway   *gw.Gateway
	bluetooth bt.Client
	config    config.Infra
	logLevel  int
}

var log = _log.Tag("infra")

func New(localID id.Identity, cfg config.Infra, querier Querier, rootDir string) (*Infra, error) {
	var i = &Infra{
		localID:  localID,
		Querier:  querier,
		rootDir:  rootDir,
		gateways: make([]infra.AddrSpec, 0),
		networks: make(map[string]infra.Network),
		config:   cfg,
	}

	// static gateways
	if err := i.setupStaticGateways(i.config); err != nil {
		log.Error("setupStaticGateways: %s", err)
	}

	// setup networks
	if err := i.setupNetworks(i.config); err != nil {
		log.Error("setupNetworks: %s", err)
	}

	return i, nil
}

func (i *Infra) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	list := make([]string, 0, len(i.networks))
	for n := range i.networks {
		list = append(list, n)
	}
	log.Log("enabled networks: %s", strings.Join(list, " "))

	for name, net := range i.networks {
		name, net := name, net
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := net.Run(ctx); err != nil {
				log.Error("network %s error: %s", log.Em(name), err.Error())
			} else {
				log.Log("network %s done", log.Em(name))
			}
		}()
	}

	wg.Wait()

	return nil
}

func (i *Infra) Networks() <-chan infra.Network {
	ch := make(chan infra.Network, len(i.networks))
	for _, n := range i.networks {
		ch <- n
	}
	close(ch)
	return ch
}

func (i *Infra) Addresses() []infra.AddrSpec {
	list := make([]infra.AddrSpec, 0)

	type addrLister interface {
		Addresses() []infra.AddrSpec
	}

	// collect addresses from all networks that support it
	for _, net := range i.networks {
		if lister, ok := net.(addrLister); ok {
			list = append(list, lister.Addresses()...)
		}
	}

	// append gateways
	list = append(list, i.gateways...)

	return list
}

func (i *Infra) Gateway() *gw.Gateway {
	return i.gateway
}

func (i *Infra) Bluetooth() bt.Client {
	return i.bluetooth
}

func (i *Infra) Tor() *tor.Tor {
	return i.tor
}

func (i *Infra) Inet() *inet.Inet {
	return i.inet
}

func (i *Infra) setupStaticGateways(cfg config.Infra) error {
	for _, gate := range cfg.Gateways {
		gateID, err := id.ParsePublicKeyHex(gate)
		if err != nil {
			log.Error("error parsing gateway %s: %s", gate, err.Error())
			continue
		}

		i.gateways = append(i.gateways, infra.AddrSpec{
			Addr:   gw.NewAddr(gateID, i.localID),
			Global: true,
		})
	}
	return nil
}
