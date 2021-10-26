package url

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/storage/types"
)

type config struct {
	Path string `mapstructure:"path"`
}

func (f urlAdapter) Create(conf *configs.Storage) types.Storage {
	return &Storage{rawConf: conf}
}
