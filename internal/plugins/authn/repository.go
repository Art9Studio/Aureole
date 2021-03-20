package authn

import (
	"aureole/configs"
	ctxTypes "aureole/context/types"
	"aureole/internal/plugins"
	"aureole/internal/plugins/authn/types"
)

var Repository = plugins.InitRepository()

// Adapter defines methods for authentication plugins
type Adapter interface {
	// Create returns desired authentication Controller depends on the given config
	Create(*configs.Authn) (types.Controller, error)
}

func InitRepository(context *ctxTypes.ProjectCtx) {
	Repository.ProjectCtx = context
}
