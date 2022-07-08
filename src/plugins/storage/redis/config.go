package redis

import "aureole/internal/configs"

type config struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Address, "localhost:6379")
}
