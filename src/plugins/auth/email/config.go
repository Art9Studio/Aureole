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
	Sender   string `mapstructure:"sender" json:"sender"`
	TmplPath string `mapstructure:"template" json:"template"`
	Exp      int    `mapstructure:"exp" json:"exp"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Exp, 600)
}
