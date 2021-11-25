package yubikey

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/2fa/types"
)

type (
	config struct {
	}
)

func (yubikeyAdapter) Create(conf *configs.SecondFactor) types.SecondFactor {
	return &yubikey{rawConf: conf}
}
