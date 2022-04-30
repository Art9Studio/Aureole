package file

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	Path string `mapstructure:"path"`
}

func (pluginCreator) Create(conf configs.PluginConfig) plugins.Plugin {
	return &storage{rawConf: conf}
}
