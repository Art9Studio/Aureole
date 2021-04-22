package storage

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/core"
	"aureole/internal/plugins/storage/types"
)

var Repository = core.CreateRepository()

// Adapter defines methods for storage plugins
type Adapter interface {
	//Create returns desired pwHasher depends on the given config
	Create(*configs.Storage) types.Storage
}
