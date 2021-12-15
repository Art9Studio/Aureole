package jwk

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
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
	RetriesNum      int    `mapstructure:"retries_num"`
	RetryInterval   int    `mapstructure:"retry_interval"`
	PathPrefix      string
}

func (jwkAdapter) Create(conf *configs.CryptoKey) plugins.CryptoKey {
	return &jwk{rawConf: conf}
}
