package email

import (
	"aureole/internal/configs"
	authnT "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		Sender     string `mapstructure:"sender"`
		Template   string `mapstructure:"template"`
		Exp        int    `mapstructure:"exp"`
		PathPrefix string
		SendUrl    string
		ConfirmUrl string
	}
)

func (emailAdapter) Create(conf *configs.Authn) authnT.Authenticator {
	return &email{rawConf: conf}
}
