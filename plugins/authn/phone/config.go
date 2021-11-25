package phone

import (
	"aureole/internal/configs"
	authnT "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		Hasher      string `mapstructure:"hasher"`
		MaxAttempts int    `mapstructure:"max_attempts"`
		Sender      string `mapstructure:"sender"`
		Template    string `mapstructure:"template"`
		Otp         otp    `mapstructure:"otp"`
		PathPrefix  string
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

func (phoneAdapter) Create(conf *configs.Authn) authnT.Authenticator {
	return &phone{rawConf: conf}
}
