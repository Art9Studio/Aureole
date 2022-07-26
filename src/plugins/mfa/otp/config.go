package authenticator

import (
	"aureole/internal/configs"
)

const (
	getQRUrl        = "/send"
	getScratchesUrl = "/scratch"
)

type config struct {
	Alg           string `mapstructure:"alg" json:"alg"`
	Iss           string `mapstructure:"iss" json:"iss"`
	WindowSize    int    `mapstructure:"window_size" json:"window_size"`
	DisallowReuse bool   `mapstructure:"disallow_reuse" json:"disallow_reuse"`
	MaxAttempts   int    `mapstructure:"max_attempts" json:"max_attempts"`
	ScratchCode   struct {
		Num      int    `mapstructure:"num" json:"num"`
		Alphabet string `mapstructure:"alphabet" json:"alphabet"`
	} `mapstructure:"scratch_code" json:"scratch_code"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Alg, "totp")
	configs.SetDefault(&c.Iss, "Aureole")
	configs.SetDefault(&c.WindowSize, 10)
	configs.SetDefault(&c.MaxAttempts, 3)
	configs.SetDefault(&c.ScratchCode.Num, 5)
	configs.SetDefault(&c.ScratchCode.Alphabet, "alphanum")
}
