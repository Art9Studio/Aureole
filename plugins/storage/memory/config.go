package memory

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/storage/types"
)

type config struct {
	Size int `mapstructure:"size"`
}

func (memoryAdapter) Create(conf *configs.Storage) types.Storage {
	return &Storage{rawConf: conf}
}
