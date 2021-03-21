package pwhasher

import (
	"aureole/configs"
	ctxTypes "aureole/context/types"
	"aureole/internal/plugins"
	"aureole/internal/plugins/pwhasher/types"
)

var Repository = plugins.InitRepository()

// Adapter defines methods for pwhasher plugins
type Adapter interface {
	//Create returns desired PwHasher depends on the given config
	Create(*configs.PwHasher) (types.PwHasher, error)
}

func InitRepository(context *ctxTypes.ProjectCtx) {
	Repository.ProjectCtx = context
}
