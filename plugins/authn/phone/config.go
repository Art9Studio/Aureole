package phone

import (
	"aureole/internal/configs"
	authnTypes "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		Hasher       string    `mapstructure:"hasher"`
		Collection   string    `mapstructure:"collection"`
		Storage      string    `mapstructure:"storage"`
		Path         string    `mapstructure:"path"`
		Verification verifConf `mapstructure:"verification"`
	}

	verifConf struct {
		Path        string
		ResendUrl   string
		Collection  string `mapstructure:"collection"`
		MaxAttempts int    `mapstructure:"max_attempts"`
		Sender      string `mapstructure:"sender"`
		Template    string `mapstructure:"template"`
		Otp         otp    `mapstructure:"otp"`
	}

	otp struct {
		Length   int    `mapstructure:"length"`
		Alphabet string `mapstructure:"alphabet"`
		Prefix   string `mapstructure:"prefix"`
		Exp      int    `mapstructure:"exp"`
	}
)

func (p phoneAdapter) Create(conf *configs.Authn) authnTypes.Authenticator {
	return &phone{rawConf: conf}
}
