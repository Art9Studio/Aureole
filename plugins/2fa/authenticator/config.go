package authenticator

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const (
	getQRUrl        = "/2fa/google-authenticator/send"
	getScratchesUrl = "/2fa/google-authenticator/scratch"
)

type (
	config struct {
		Alg           string      `mapstructure:"alg"`
		Iss           string      `mapstructure:"iss"`
		WindowSize    int         `mapstructure:"window_size"`
		DisallowReuse bool        `mapstructure:"disallow_reuse"`
		MaxAttempts   int         `mapstructure:"max_attempts"`
		ScratchCode   scratchCode `mapstructure:"scratch_code"`
	}

	scratchCode struct {
		Num      int    `mapstructure:"num"`
		Alphabet string `mapstructure:"alphabet"`
	}
)

func (gauthAdapter) Create(conf *configs.SecondFactor) plugins.SecondFactor {
	return &gauth{rawConf: conf}
}
