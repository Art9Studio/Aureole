package sms

import "aureole/internal/configs"

func (c *config) setDefaults() {
	c.PathPrefix = "/2fa/sms"
	c.SendUrl = "/send"
	c.VerifyUrl = "/verify"
	c.ResendUrl = "/resend"
	configs.SetDefault(&c.MaxAttempts, 3)
	c.Otp.setDefaults()
}

func (o *otp) setDefaults() {
	configs.SetDefault(&o.Length, 6)
	configs.SetDefault(&o.Alphabet, "alphanum")
	configs.SetDefault(&o.Exp, 300)
}
