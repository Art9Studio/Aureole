package kstorage

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/core"
	"aureole/internal/plugins/kstorage/types"
)

var Repository = core.CreateRepository()

type Adapter interface {
	Create(*configs.KeyStorage) types.KeyStorage
}
