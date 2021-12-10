package sms

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/2fa/types"
)

type (
	config struct {
		Sender      string `mapstructure:"sender"`
		Template    string `mapstructure:"template"`
		MaxAttempts int    `mapstructure:"max_attempts"`
		Otp         otp    `mapstructure:"otp"`
		PathPrefix  string
		SendUrl     string
		VerifyUrl   string
		ResendUrl   string
	}

	otp struct {
		Length   int    `mapstructure:"length"`
		Alphabet string `mapstructure:"alphabet"`
		Exp      int    `mapstructure:"exp"`
	}
)

func (smsAdapter) Create(conf *configs.SecondFactor) types.SecondFactor {
	return &sms{rawConf: conf}
}
