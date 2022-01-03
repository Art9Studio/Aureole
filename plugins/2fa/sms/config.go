package sms

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const resendUrl = "/2fa/sms/resend"

type (
	config struct {
		Sender      string `mapstructure:"sender"`
		Template    string `mapstructure:"template"`
		MaxAttempts int    `mapstructure:"max_attempts"`
		Otp         otp    `mapstructure:"otp"`
	}

	otp struct {
		Length   int    `mapstructure:"length"`
		Alphabet string `mapstructure:"alphabet"`
		Exp      int    `mapstructure:"exp"`
	}
)

func (smsAdapter) Create(conf *configs.SecondFactor) plugins.SecondFactor {
	return &sms{rawConf: conf}
}
