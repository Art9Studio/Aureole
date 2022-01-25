package url

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	Path string `mapstructure:"path"`
}

func (urlAdapter) Create(conf *configs.CryptoStorage) plugins.CryptoStorage {
	return &storage{rawConf: conf}
}
