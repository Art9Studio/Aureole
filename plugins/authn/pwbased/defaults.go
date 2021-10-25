package pwbased

import (
	"aureole/internal/configs"
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.CompatHashers, []string{})
	c.Login.setDefaults()
	c.Register.setDefaults()
	c.Reset.setDefaults()
	c.Verif.setDefaults()
}

func (l *login) setDefaults() {
	l.Path = "/login"
}

func (r *register) setDefaults() {
	r.Path = "/register"
}

func (c *resetConf) setDefaults() {
	c.Path = "/reset-password"
	c.ConfirmUrl = "/reset-password/confirm"
	configs.SetDefault(&c.Token.Exp, 3600)
	configs.SetDefault(&c.Token.HashFunc, "sha256")
}

func (c *verifConf) setDefaults() {
	c.Path = "/verify-email"
	c.ConfirmUrl = "/verify-email/confirm"
	configs.SetDefault(&c.Token.Exp, 3600)
	configs.SetDefault(&c.Token.HashFunc, "sha256")
}
