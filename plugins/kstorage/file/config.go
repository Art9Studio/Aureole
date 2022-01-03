package file

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	Path string `mapstructure:"path"`
}

func (fileAdapter) Create(conf *configs.KeyStorage) plugins.KeyStorage {
	return &storage{rawConf: conf}
}
