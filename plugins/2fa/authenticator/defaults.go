package authenticator

import "aureole/internal/configs"

func (c *config) setDefaults() {
	c.PathPrefix = "/2fa/google-authenticator"
	c.GetQRUrl = "/send"
	c.VerifyUrl = "/verify"
	c.GetScratchesUrl = "/scratch"
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
