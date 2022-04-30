package sms

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const (
	resendUrl   = "/2fa/sms/resend"
	defaultTmpl = "Your second factor code: {{.otp}}"
)

type config struct {
	Sender      string `mapstructure:"sender"`
	TmplPath    string `mapstructure:"template"`
	MaxAttempts int    `mapstructure:"max_attempts"`
	Otp         struct {
		Length   int    `mapstructure:"length"`
		Alphabet string `mapstructure:"alphabet"`
		Exp      int    `mapstructure:"exp"`
	} `mapstructure:"otp"`
}

func (smsPlugin) Create(conf configs.PluginConfig) plugins.Plugin {
	return &sms{rawConf: conf}
}
