package sms

import (
	"aureole/configs"
)

const (
	resendUrl   = "/resend_token"
	initMFA     = "/init"
	send        = "/send"
	defaultTmpl = "Your second factor code: {{.otp}}"
)

type config struct {
	Sender      string `mapstructure:"sender" json:"sender"`
	TmplPath    string `mapstructure:"template" json:"template"`
	MaxAttempts int    `mapstructure:"max_attempts" json:"max_attempts"`
	Otp         struct {
		Length   int    `mapstructure:"length" json:"length"`
		Alphabet string `mapstructure:"alphabet" json:"alphabet"`
		Exp      int    `mapstructure:"exp" json:"exp"`
		Prefix   string `mapstructure:"prefix" json:"prefix"`
		Postfix  string `mapstructure:"postfix" json:"postfix"`
	} `mapstructure:"otp" json:"otp"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.MaxAttempts, 3)
	configs.SetDefault(&c.Otp.Length, 6)
	configs.SetDefault(&c.Otp.Alphabet, "alphanum")
	configs.SetDefault(&c.Otp.Exp, 300)
}
