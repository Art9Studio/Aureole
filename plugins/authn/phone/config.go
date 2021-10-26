package phone

import (
	"aureole/internal/configs"
	authnTypes "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		Hasher      string `mapstructure:"hasher"`
		MaxAttempts int    `mapstructure:"max_attempts"`
		Sender      string `mapstructure:"sender"`
		Template    string `mapstructure:"template"`
		Otp         otp    `mapstructure:"otp"`
		SendUrl     string
		ConfirmUrl  string
		ResendUrl   string
	}

	otp struct {
		Length   int    `mapstructure:"length"`
		Alphabet string `mapstructure:"alphabet"`
		Prefix   string `mapstructure:"prefix"`
		Postfix  string `mapstructure:"postfix"`
		Exp      int    `mapstructure:"exp"`
	}
)

func (p phoneAdapter) Create(conf *configs.Authn) authnTypes.Authenticator {
	return &phone{rawConf: conf}
}
