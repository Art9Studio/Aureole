package authn

import (
	"gouth/adapters/authn/types"
	"gouth/config"
	contextTypes "gouth/context/types"
)

func New(conf *config.AuthnConfig, projectCtx *contextTypes.ProjectCtx) (types.Controller, error) {
	adapter, err := GetAdapter(conf.TypeName)
	if err != nil {
		return nil, err
	}

	return adapter.GetAuthnController(conf.PathPrefix, &conf.Config, projectCtx)
}
