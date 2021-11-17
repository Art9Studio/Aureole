package url

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/kstorage/types"
)

type config struct {
	Path string `mapstructure:"path"`
}

func (urlAdapter) Create(conf *configs.KeyStorage) types.KeyStorage {
	return &Storage{rawConf: conf}
}
