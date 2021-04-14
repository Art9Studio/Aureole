package jwk

import (
	"aureole/configs"
	"aureole/internal/plugins/cryptokey/types"
)

type config struct {
	Path string `mapstructure:"path"`
}

func (j jwkAdapter) Create(conf *configs.CryptoKey) types.CryptoKey {
	return &Jwk{rawConf: conf}
}
