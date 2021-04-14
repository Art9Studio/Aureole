package authz

import (
	"aureole/configs"
	ctxTypes "aureole/context/types"
	"aureole/internal/plugins"
	"aureole/internal/plugins/authz/types"
)

var Repository = plugins.InitRepository()

// Adapter defines methods for authorization plugins
type Adapter interface {
	// Create returns desired authorization depends on the given config
	Create(*configs.Authz) types.Authorizer
}

func InitRepository(context *ctxTypes.ProjectCtx) {
	Repository.ProjectCtx = context
}
