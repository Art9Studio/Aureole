package email

import (
	"aureole/internal/configs"
)

const (
	sendUrl     = "/email/send"
	loginUrl    = "/email/login"
	defaultTmpl = "Click and confirm that you want to sign in: {{.link}}"
)

type config struct {
	Sender   string `mapstructure:"sender"`
	TmplPath string `mapstructure:"template"`
	Exp      int    `mapstructure:"exp"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Exp, 600)
}
