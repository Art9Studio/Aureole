package authn

import (
	"aureole/configs"
	ctxTypes "aureole/context/types"
	"aureole/plugins"
	"aureole/plugins/authn/types"
)

var Repository = plugins.InitRepository()

// Adapter defines methods for authentication plugins
type Adapter interface {
	// GetAuthnController returns desired authentication Controller depneds on the given config
	Create(*configs.Authn, *ctxTypes.ProjectCtx) (types.Controller, error)
}

func InitRepository(context *ctxTypes.ProjectCtx) {
	Repository.ProjectCtx = context
}
