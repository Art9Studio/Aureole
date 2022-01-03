package authenticator

import "aureole/internal/configs"

func (c *config) setDefaults() {
	configs.SetDefault(&c.Alg, "totp")
	configs.SetDefault(&c.Iss, "Aureole")
	configs.SetDefault(&c.WindowSize, 10)
	configs.SetDefault(&c.MaxAttempts, 3)
	c.ScratchCode.setDefaults()
}

func (s *scratchCode) setDefaults() {
	configs.SetDefault(&s.Num, 5)
	configs.SetDefault(&s.Alphabet, "alphanum")
}
