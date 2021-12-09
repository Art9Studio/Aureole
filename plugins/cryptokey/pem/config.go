package pem

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	Alg             string `mapstructure:"alg"`
	Storage         string `mapstructure:"storage"`
	RefreshInterval int    `mapstructure:"refresh_interval"`
	PathPrefix      string
}

func (pemAdapter) Create(conf *configs.CryptoKey) plugins.CryptoKey {
	return &pem{rawConf: conf}
}
