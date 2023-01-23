package config

import (
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
)

// Infra holds configs for individual infrastructural networks
type Infra struct {
	Networks    []string    `yaml:"networks"`
	Gateways    []string    `yaml:"gateways"`
	StickyNodes []string    `yaml:"sticky_nodes"`
	Inet        inet.Config `yaml:"inet"`
	Tor         tor.Config  `yaml:"tor"`
	Gw          gw.Config   `yaml:"gw"`
}

func (i Infra) IsNetworkEnabled(network string) bool {
	if len(i.Networks) == 0 {
		return true
	}

	for _, n := range i.Networks {
		if n == network {
			return true
		}
	}

	return false
}
