package yubikey

import (
	"aureole/internal/configs"
	"aureole/internal/core"
)

type (
	config struct {
	}
)

func (yubikeyPlugin) Create(conf configs.PluginConfig) core.MFA {
	return &yubikey{rawConf: conf}
}
