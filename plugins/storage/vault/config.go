package vault

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/storage/types"
)

type config struct {
	Path    string `mapstructure:"path"`
	Token   string `mapstructure:"token"`
	Address string `mapstructure:"address"`
}

func (v vaultAdapter) Create(conf *configs.Storage) types.Storage {
	return &Storage{rawConf: conf}
}
