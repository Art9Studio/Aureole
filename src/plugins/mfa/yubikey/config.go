package yubikey

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type (
	config struct {
	}
)

func (yubikeyPlugin) Create(conf configs.PluginConfig) plugins.Plugin {
	return &yubikey{rawConf: conf}
}
