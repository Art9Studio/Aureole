package jwk

import "aureole/internal/configs"

func (c *config) setDefaults() {
	configs.SetDefault(&c.Kid, "SHA-256")
}