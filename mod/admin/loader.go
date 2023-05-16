package admin

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "admin"

type Loader struct{}

func (Loader) Load(node modules.Node, configStore config.Store) (modules.Module, error) {
	mod := &Module{
		config:   defaultConfig,
		node:     node,
		commands: make(map[string]Command),
		log:      log.Tag(ModuleName),
	}

	configStore.LoadYAML(ModuleName, &mod.config)

	mod.AddCommand("help", &CmdHelp{mod: mod})
	mod.AddCommand("tracker", &CmdTracker{mod: mod})
	mod.AddCommand("net", &CmdNet{mod: mod})

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
