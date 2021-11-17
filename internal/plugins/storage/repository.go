package storage

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/core"
	"aureole/internal/plugins/storage/types"
)

var Repository = core.CreateRepository()

type Adapter interface {
	Create(*configs.Storage) types.Storage
}
