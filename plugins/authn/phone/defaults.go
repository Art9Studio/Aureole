package phone

import (
	"aureole/internal/configs"
)

func (c *config) setDefaults() {
	c.Path = "/send"
	c.Verification.setDefaults()
}

func (v *verifConf) setDefaults() {
	v.Path = "/login"
	v.ResendUrl = "/resend"
	configs.SetDefault(&v.MaxAttempts, 3)
	v.Otp.setDefaults()
}
func (c *otp) setDefaults() {
	configs.SetDefault(&c.Length, 1)
	configs.SetDefault(&c.Alphabet, "1234567890")
	configs.SetDefault(&c.Exp, 300)
}
