package authenticator

import (
	"aureole/internal/configs"
)

const (
	getQRUrl        = "/2fa/otp/send"
	getScratchesUrl = "/2fa/otp/scratch"
)

type config struct {
	Alg           string `mapstructure:"alg"`
	Iss           string `mapstructure:"iss"`
	WindowSize    int    `mapstructure:"window_size"`
	DisallowReuse bool   `mapstructure:"disallow_reuse"`
	MaxAttempts   int    `mapstructure:"max_attempts"`
	ScratchCode   struct {
		Num      int    `mapstructure:"num"`
		Alphabet string `mapstructure:"alphabet"`
	} `mapstructure:"scratch_code"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Alg, "totp")
	configs.SetDefault(&c.Iss, "Aureole")
	configs.SetDefault(&c.WindowSize, 10)
	configs.SetDefault(&c.MaxAttempts, 3)
	configs.SetDefault(&c.ScratchCode.Num, 5)
	configs.SetDefault(&c.ScratchCode.Alphabet, "alphanum")
}
