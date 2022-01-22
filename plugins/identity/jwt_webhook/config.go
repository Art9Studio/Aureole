package jwt_webhook

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	Address       string            `mapstructure:"address"`
	RetriesNum    uint              `mapstructure:"retries_num"`
	RetryInterval int               `mapstructure:"retry_interval"`
	Timeout       int               `mapstructure:"timeout"`
	Headers       map[string]string `mapstructure:"headers"`
}

func (adapter) Create(conf *configs.IDManager) plugins.IDManager {
	return &manager{rawConf: conf}
}
