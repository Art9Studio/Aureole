package authenticator

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/2fa/types"
)

type (
	config struct {
		Alg             string      `mapstructure:"alg"`
		Iss             string      `mapstructure:"iss"`
		WindowSize      int         `mapstructure:"window_size"`
		DisallowReuse   bool        `mapstructure:"disallow_reuse"`
		MaxAttempts     int         `mapstructure:"max_attempts"`
		ScratchCode     scratchCode `mapstructure:"scratch_code"`
		PathPrefix      string
		GetQRUrl        string
		VerifyUrl       string
		GetScratchesUrl string
	}

	scratchCode struct {
		Num      int    `mapstructure:"num"`
		Alphabet string `mapstructure:"alphabet"`
	}
)

func (gauthAdapter) Create(conf *configs.SecondFactor) types.SecondFactor {
	return &gauth{rawConf: conf}
}
