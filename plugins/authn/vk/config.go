package vk

import (
	"aureole/internal/configs"
	authnT "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		ClientId     string   `mapstructure:"client_id"`
		ClientSecret string   `mapstructure:"client_secret"`
		Scopes       []string `mapstructure:"scopes"`
		Fields       []string `mapstructure:"fields"`
		PathPrefix   string
		RedirectUri  string
	}
)

func (vkAdapter) Create(conf *configs.Authn) authnT.Authenticator {
	return &vk{rawConf: conf}
}
