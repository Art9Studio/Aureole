package authz

import (
	"aureole/configs"
	"aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/core"
)

var Repository = core.CreateRepository()

// Adapter defines methods for authorization plugins
type Adapter interface {
	// Create returns desired authorization depends on the given config
	Create(*configs.Authz) types.Authorizer
}

func InitRepository(api *core.PluginsApi) {
	Repository.PluginsApi = api
}
