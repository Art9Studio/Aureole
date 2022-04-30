package etcd

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	Endpoints   []string `mapstructure:"endpoints"`
	Timeout     float32  `mapstructure:"timeout"`
	DialTimeout float32  `mapstructure:"dial_timeout"`
}

func (etcdPlugin) Create(conf configs.PluginConfig) plugins.Plugin {
	return &etcd{rawConf: conf}
}
