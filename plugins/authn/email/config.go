package email

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
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

func (emailAdapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &email{rawConf: conf}
}
