package phone

import (
	"aureole/internal/configs"
)

func (c *config) setDefaults() {
	c.PathPrefix = "/" + AdapterName
	c.SendUrl = "/send"
	c.ConfirmUrl = "/login"
	c.ResendUrl = "/resend"
	configs.SetDefault(&c.MaxAttempts, 3)
	c.Otp.setDefaults()
}

func (c *otp) setDefaults() {
	configs.SetDefault(&c.Length, 6)
	configs.SetDefault(&c.Alphabet, "1234567890")
	configs.SetDefault(&c.Exp, 300)
}
