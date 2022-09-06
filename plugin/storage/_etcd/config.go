package etcd

import (
	"aureole/configs"
	"aureole/internal/core"
)

type config struct {
	Endpoints   []string `mapstructure:"endpoints" json:"endpoints"`
	Timeout     float32  `mapstructure:"timeout" json:"timeout"`
	DialTimeout float32  `mapstructure:"dial_timeout" json:"dial_timeout"`
}

func (etcdPlugin) Create(conf configs.PluginConfig) core.Storage {
	return &etcd{rawConf: conf}
}
