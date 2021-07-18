package pem

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/cryptokey/types"
)

type config struct {
	Alg  string `mapstructure:"alg"`
	Path string `mapstructure:"path"`
}

func (p pemAdapter) Create(conf *configs.CryptoKey) types.CryptoKey {
	return &Pem{rawConf: conf}
}
