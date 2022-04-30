package urls

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const getUrlsPath = "urls"

// Create returns urls hasher with the given settings
func (pluginCreator) Create(conf configs.PluginConfig) plugins.Plugin {
	return &urls{rawConf: conf}
}
