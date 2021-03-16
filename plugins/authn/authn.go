package authn

import (
	"aureole/configs"
	"aureole/plugins/authn/types"
)

// New returns desired Controller depends on the given config
func New(conf *configs.Authn) (types.Controller, error) {
	adapter, err := GetAdapter(conf.Type)
	if err != nil {
		return nil, err
	}

	return adapter.Get(conf, projectCtx)
}
