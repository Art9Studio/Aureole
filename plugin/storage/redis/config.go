package redis

import (
	"aureole/configs"
)

type config struct {
	Address  string `mapstructure:"address" json:"address"`
	Password string `mapstructure:"password" json:"password"`
	DB       int    `mapstructure:"db" json:"db"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Address, "localhost:6379")
}
