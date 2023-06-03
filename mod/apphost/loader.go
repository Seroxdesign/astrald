package apphost

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"net"
)

const ModuleName = "apphost"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store) (modules.Module, error) {
	var err error

	mod := &Module{
		config:    defaultConfig,
		node:      node,
		listeners: make([]net.Listener, 0),
		tokens:    make(map[string]id.Identity, 0),
		execs:     []*Exec{},
		log:       log.Tag(ModuleName),
	}

	_ = assets.LoadYAML(ModuleName, &mod.config)

	mod.keys, err = assets.KeyStore()
	if err != nil {
		return nil, err
	}

	adm, err := modules.Find[*admin.Module](node.Modules())
	if err == nil {
		_ = adm.AddCommand("apphost", &Admin{mod: mod})
	}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
