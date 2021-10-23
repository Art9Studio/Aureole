package email

import (
	"aureole/internal/configs"
	authnTypes "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		Collection string `mapstructure:"collection"`
		Storage    string `mapstructure:"storage"`
		Sender     string `mapstructure:"sender"`
		Template   string `mapstructure:"template"`
		Exp        int    `mapstructure:"exp"`
		SendUrl    string
		ConfirmUrl string
	}
)

func (e emailAdapter) Create(conf *configs.Authn) authnTypes.Authenticator {
	return &email{rawConf: conf}
}
