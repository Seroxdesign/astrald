package tor

import (
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "tor"

type API interface {
	infra.Dialer
	infra.Unpacker
	infra.Parser
	infra.EndpointLister
}

func Load(node modules.Node) (API, error) {
	api, ok := node.Modules().Find(ModuleName).(API)
	if !ok {
		return nil, modules.ErrNotFound
	}
	return api, nil
}
