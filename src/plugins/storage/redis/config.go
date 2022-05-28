package redis

import (
	"aureole/internal/configs"
	"aureole/internal/core"
)

type config struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (redisPlugin) Create(conf configs.PluginConfig) core.Storage {
	return &redis{rawConf: conf}
}
