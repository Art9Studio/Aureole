package standard

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	DBUrl string `mapstructure:"db_url"`
}

func (plugin) Create(conf configs.PluginConfig) plugins.Plugin {
	return &manager{rawConf: conf}
}
