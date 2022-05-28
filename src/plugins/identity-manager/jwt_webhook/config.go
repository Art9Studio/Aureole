package jwt_webhook

import (
	"aureole/internal/configs"
	"aureole/internal/plugin"
)

type config struct {
	Address       string            `mapstructure:"address"`
	RetriesNum    int               `mapstructure:"retries_num"`
	RetryInterval int               `mapstructure:"retry_interval"`
	Timeout       int               `mapstructure:"timeout"`
	Headers       map[string]string `mapstructure:"headers"`
}

func (plugin) Create(conf configs.PluginConfig) plugin.Plugin {
	return &manager{rawConf: conf}
}
