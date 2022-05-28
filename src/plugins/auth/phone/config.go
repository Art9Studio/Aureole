package phone

import (
	"aureole/internal/configs"
)

const (
	sendUrl     = "/phone/send"
	resendUrl   = "/phone/resend"
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

func (c *config) setDefaults() {
	configs.SetDefault(&c.MaxAttempts, 3)
	configs.SetDefault(&c.Otp.Length, 6)
	configs.SetDefault(&c.Otp.Alphabet, "num")
	configs.SetDefault(&c.Otp.Exp, 300)
}
