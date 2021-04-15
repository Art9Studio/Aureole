package storage

import (
	"aureole/configs"
	ctxTypes "aureole/context/types"
	"aureole/internal/plugins"
	"aureole/internal/plugins/storage/types"
)

var Repository = plugins.InitRepository()

// Adapter defines methods for storage plugins
type Adapter interface {
	//Create returns desired pwHasher depends on the given config
	Create(*configs.Storage) types.Storage
}

func InitRepository(context *ctxTypes.ProjectCtx) {
	Repository.ProjectCtx = context
}
