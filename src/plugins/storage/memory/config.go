package memory

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	Size int `mapstructure:"size"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Size, 128)
}

func (pluginCreator) Create(conf configs.PluginConfig) plugins.Plugin {
	return &memory{rawConf: conf}
}
