package etcd

import (
	"aureole/internal/configs"
)

type config struct {
	Endpoints   []string `mapstructure:"endpoints"`
	Timeout     float32  `mapstructure:"timeout"`
	DialTimeout float32  `mapstructure:"dial_timeout"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Endpoints, []string{"localhost:2379"})
	configs.SetDefault(&c.Timeout, 0.2)
	configs.SetDefault(&c.DialTimeout, 2)
}
