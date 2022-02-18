package phone

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const (
	sendUrl     = "/phone/send"
	resendUrl   = "/phone/resendOTP"
	defaultTmpl = "Your verification code: {{.otp}}"
)

type config struct {
	MaxAttempts int    `mapstructure:"max_attempts"`
	Sender      string `mapstructure:"sender"`
	TmplPath    string `mapstructure:"template"`
	Otp         struct {
		Length   int    `mapstructure:"length"`
		Alphabet string `mapstructure:"alphabet"`
		Prefix   string `mapstructure:"prefix"`
		Postfix  string `mapstructure:"postfix"`
		Exp      int    `mapstructure:"exp"`
	} `mapstructure:"otp"`
}

func (phoneAdapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &phone{rawConf: conf}
}
