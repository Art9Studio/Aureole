package email

import (
	"aureole/internal/configs"
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.Exp, 600)
}
