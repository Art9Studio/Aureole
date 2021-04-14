package pwhasher

import (
	"aureole/configs"
	"aureole/internal/plugins/core"
	"aureole/internal/plugins/pwhasher/types"
)

var Repository = core.InitRepository()

// Adapter defines methods for pwhasher plugins
type Adapter interface {
	//Create returns desired pwHasher depends on the given config
	Create(*configs.PwHasher) types.PwHasher
}

func InitRepository(api *core.PluginApi) {
	Repository.PluginApi = api
}
