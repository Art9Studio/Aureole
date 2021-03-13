package authn

import (
	"gouth/adapters/authn/types"
	"gouth/configs"
)

func New(conf *configs.AuthnConfig) (types.Controller, error) {
	adapter, err := GetAdapter(conf.Type)
	if err != nil {
		return nil, err
	}

	return adapter.GetAuthnController(conf.PathPrefix, &conf.Config, projectCtx)
}
