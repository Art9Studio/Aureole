package standard

import (
	"aureole/internal/configs"
	"aureole/internal/plugin"
)

type config struct {
	DBUrl string `mapstructure:"db_url"`
}

func (plugin) Create(conf configs.PluginConfig) plugin.Plugin {
	return &manager{rawConf: conf}
}
