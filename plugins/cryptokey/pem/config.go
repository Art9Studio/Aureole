package pem

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/cryptokey/types"
)

type config struct {
	Alg             string `mapstructure:"alg"`
	Storage         string `mapstructure:"storage"`
	RefreshInterval int    `mapstructure:"refresh_interval"`
}

func (p pemAdapter) Create(conf *configs.CryptoKey) types.CryptoKey {
	return &Pem{rawConf: conf}
}
