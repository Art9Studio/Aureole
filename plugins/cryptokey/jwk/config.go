package jwk

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/cryptokey/types"
)

type config struct {
	Kty             string `mapstructure:"kty"`
	Alg             string `mapstructure:"alg"`
	Curve           string `mapstructure:"curve"`
	Size            int    `mapstructure:"size"`
	Kid             string `mapstructure:"kid"`
	Storage         string `mapstructure:"storage"`
	RefreshInterval int    `mapstructure:"refresh_interval"`
}

func (j jwkAdapter) Create(conf *configs.CryptoKey) types.CryptoKey {
	return &Jwk{rawConf: conf}
}
