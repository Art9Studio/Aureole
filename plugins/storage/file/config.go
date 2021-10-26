package file

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/storage/types"
)

type config struct {
	Path string `mapstructure:"path"`
}

func (f fileAdapter) Create(conf *configs.Storage) types.Storage {
	return &Storage{rawConf: conf}
}
