package jwt_webhook

import "aureole/internal/configs"

func (c *config) setDefaults() {
	configs.SetDefault(&c.RetriesNum, 1)
	configs.SetDefault(&c.RetryInterval, 0.5)
	configs.SetDefault(&c.Timeout, 5)
}
