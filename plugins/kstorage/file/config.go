package file

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/kstorage/types"
)

type config struct {
	Path string `mapstructure:"path"`
}

func (fileAdapter) Create(conf *configs.KeyStorage) types.KeyStorage {
	return &storage{rawConf: conf}
}
