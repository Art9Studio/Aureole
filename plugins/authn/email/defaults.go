package email

import (
	"aureole/internal/configs"
)

func (c *config) setDefaults() {
	c.Path = "/send"
	c.Link.setDefaults()
}

func (t *token) setDefaults() {
	configs.SetDefault(&t.Exp, 600)
	configs.SetDefault(&t.HashFunc, "sha256")
}

func (m *magicLinkConf) setDefaults() {
	m.Path = "/login"
	m.Token.setDefaults()
}
