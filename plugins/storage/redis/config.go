package redis

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (redisAdapter) Create(conf *configs.Storage) plugins.Storage {
	return &redis{rawConf: conf}
}
