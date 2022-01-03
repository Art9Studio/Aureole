package phone

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const (
	sendUrl   = "/phone/send"
	resendUrl = "/phone/resendOTP"
)

type (
	config struct {
		Hasher      string `mapstructure:"hasher"`
		MaxAttempts int    `mapstructure:"max_attempts"`
		Sender      string `mapstructure:"sender"`
		Template    string `mapstructure:"template"`
		Otp         otp    `mapstructure:"otp"`
	}

	otp struct {
		Length   int    `mapstructure:"length"`
		Alphabet string `mapstructure:"alphabet"`
		Prefix   string `mapstructure:"prefix"`
		Postfix  string `mapstructure:"postfix"`
		Exp      int    `mapstructure:"exp"`
	}
)

func (phoneAdapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &phone{rawConf: conf}
}
