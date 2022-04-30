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

func (redisPlugin) Create(conf configs.PluginConfig) plugins.Plugin {
	return &redis{rawConf: conf}
}
