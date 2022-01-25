package sms

import "aureole/internal/configs"

func (c *config) setDefaults() {
	configs.SetDefault(&c.MaxAttempts, 3)
	configs.SetDefault(&c.Otp.Length, 6)
	configs.SetDefault(&c.Otp.Alphabet, "alphanum")
	configs.SetDefault(&c.Otp.Exp, 300)
}
