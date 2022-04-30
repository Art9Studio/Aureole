package pem

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	Alg             string `mapstructure:"alg"`
	Storage         string `mapstructure:"storage"`
	RefreshInterval int    `mapstructure:"refresh_interval"`
	RetriesNum      int    `mapstructure:"retries_num"`
	RetryInterval   int    `mapstructure:"retry_interval"`
	PathPrefix      string
}

func (pemPlugin) Create(conf configs.PluginConfig) plugins.Plugin {
	return &pem{rawConf: conf}
}
