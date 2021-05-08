package authn

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/authn/types"
	"aureole/internal/plugins/core"
)

var Repository = core.CreateRepository()

// Adapter defines methods for authentication plugins
type Adapter interface {
	// Create returns desired authentication Authenticator depends on the given config
	Create(*configs.Authn) types.Authenticator
}
