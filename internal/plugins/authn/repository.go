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
	// Create returns desired authentication Authenticator depends on the given config
	Create(string, *configs.Authn) types.Authenticator
}

func InitRepository(context *ctxTypes.ProjectCtx) {
	Repository.ProjectCtx = context
}
