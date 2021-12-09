package email

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const (
	sendUrl  = "/email/send"
	loginUrl = "/email/login"
)

type (
	config struct {
		Sender   string `mapstructure:"sender"`
		Template string `mapstructure:"template"`
		Exp      int    `mapstructure:"exp"`
	}
)

func (emailAdapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &email{rawConf: conf}
}
