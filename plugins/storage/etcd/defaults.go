package etcd

import "aureole/internal/configs"

func (c *config) setDefaults() {
	configs.SetDefault(&c.Endpoints, []string{"localhost:2379"})
	configs.SetDefault(&c.Timeout, 0.2)
	configs.SetDefault(&c.DialTimeout, 2)
}
