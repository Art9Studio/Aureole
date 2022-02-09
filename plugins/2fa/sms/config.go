package sms

import (
	"aureole/internal/configs"
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

func (c *config) setDefaults() {
	configs.SetDefault(&c.MaxAttempts, 3)
	configs.SetDefault(&c.Otp.Length, 6)
	configs.SetDefault(&c.Otp.Alphabet, "alphanum")
	configs.SetDefault(&c.Otp.Exp, 300)
}
