package phone

import (
	"aureole/internal/configs"
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.MaxAttempts, 3)
	c.Otp.setDefaults()
}

func (c *otp) setDefaults() {
	configs.SetDefault(&c.Length, 6)
	configs.SetDefault(&c.Alphabet, "num")
	configs.SetDefault(&c.Exp, 300)
}
