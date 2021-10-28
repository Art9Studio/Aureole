package email

import (
	"aureole/internal/configs"
)

func (c *config) setDefaults() {
	c.SendUrl = "/send"
	c.ConfirmUrl = "/login"
	configs.SetDefault(&c.Exp, 600)
}
