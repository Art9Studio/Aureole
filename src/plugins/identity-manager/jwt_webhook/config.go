package jwt_webhook

import (
	"aureole/internal/configs"
)

type config struct {
	Address       string            `mapstructure:"address"`
	RetriesNum    int               `mapstructure:"retries_num"`
	RetryInterval int               `mapstructure:"retry_interval"`
	Timeout       int               `mapstructure:"timeout"`
	Headers       map[string]string `mapstructure:"headers"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.RetriesNum, 1)
	configs.SetDefault(&c.RetryInterval, 100)
	configs.SetDefault(&c.Timeout, 5)
}
