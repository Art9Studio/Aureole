package jwk

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/cryptokey/types"
)

type config struct {
	Kty             string `mapstructure:"kty"`
	Alg             string `mapstructure:"alg"`
	Use             string `mapstructure:"use"`
	Curve           string `mapstructure:"curve"`
	Size            int    `mapstructure:"size"`
	Kid             string `mapstructure:"kid"`
	Storage         string `mapstructure:"storage"`
	RefreshInterval int    `mapstructure:"refresh_interval"`
	PathPrefix      string
}

func (jwkAdapter) Create(conf *configs.CryptoKey) types.CryptoKey {
	return &Jwk{rawConf: conf}
}
