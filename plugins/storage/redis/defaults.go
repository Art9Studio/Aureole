package redis

import "aureole/internal/configs"

func (c *config) setDefaults() {
	configs.SetDefault(&c.Address, "localhost:6379")
}
