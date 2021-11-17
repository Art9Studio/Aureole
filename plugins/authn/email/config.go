package email

import (
	"aureole/internal/configs"
	authnTypes "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		Sender     string `mapstructure:"sender"`
		Template   string `mapstructure:"template"`
		Exp        int    `mapstructure:"exp"`
		SendUrl    string
		ConfirmUrl string
	}
)

func (emailAdapter) Create(conf *configs.Authn) authnTypes.Authenticator {
	return &email{rawConf: conf}
}
