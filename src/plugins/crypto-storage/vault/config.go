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

func (pluginCreator) Create(conf configs.PluginConfig) plugins.Plugin {
	return &storage{rawConf: conf}
}
