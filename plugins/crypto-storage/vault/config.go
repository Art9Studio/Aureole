package vault

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	Path    string `mapstructure:"path"`
	Token   string `mapstructure:"token"`
	Address string `mapstructure:"address"`
}

func (vaultAdapter) Create(conf *configs.CryptoStorage) plugins.CryptoStorage {
	return &storage{rawConf: conf}
}
