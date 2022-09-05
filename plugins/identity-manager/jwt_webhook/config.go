package jwt_webhook

import (
	"aureole/internal/configs"
)

type config struct {
	Address       string            `mapstructure:"address" json:"address"`
	RetriesNum    int               `mapstructure:"retries_num" json:"retries_num"`
	RetryInterval int               `mapstructure:"retry_interval" json:"retry_interval"`
	Timeout       int               `mapstructure:"timeout" json:"timeout"`
	Headers       map[string]string `mapstructure:"headers" json:"headers"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.RetriesNum, 1)
	configs.SetDefault(&c.RetryInterval, 100)
	configs.SetDefault(&c.Timeout, 5)
}
