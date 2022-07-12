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
	MaxAttempts int    `mapstructure:"max_attempts" json:"max_attempts"`
	Sender      string `mapstructure:"sender" json:"sender"`
	TmplPath    string `mapstructure:"template" json:"template"`
	Otp         struct {
		Length   int    `mapstructure:"length" json:"length"`
		Alphabet string `mapstructure:"alphabet" json:"alphabet"`
		Prefix   string `mapstructure:"prefix" json:"prefix"`
		Postfix  string `mapstructure:"postfix" json:"postfix"`
		Exp      int    `mapstructure:"exp" json:"exp"`
	} `mapstructure:"otp" json:"otp"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.MaxAttempts, 3)
	configs.SetDefault(&c.Otp.Length, 6)
	configs.SetDefault(&c.Otp.Alphabet, "num")
	configs.SetDefault(&c.Otp.Exp, 300)
}
