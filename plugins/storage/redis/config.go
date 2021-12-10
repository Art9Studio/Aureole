package redis

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/storage/types"
)

type config struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (redisAdapter) Create(conf *configs.Storage) types.Storage {
	return &Storage{rawConf: conf}
}
