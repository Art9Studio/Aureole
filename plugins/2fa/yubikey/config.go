package yubikey

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type (
	config struct {
	}
)

func (yubikeyAdapter) Create(conf *configs.SecondFactor) plugins.SecondFactor {
	return &yubikey{rawConf: conf}
}
