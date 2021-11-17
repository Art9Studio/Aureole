package vk

import (
	"aureole/internal/configs"
	authnTypes "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		ClientId     string   `mapstructure:"client_id"`
		ClientSecret string   `mapstructure:"client_secret"`
		Scopes       []string `mapstructure:"scopes"`
		RedirectUri  string
		Fields       []string `mapstructure:"fields"`
	}
)

func (vkAdapter) Create(conf *configs.Authn) authnTypes.Authenticator {
	return &vk{rawConf: conf}
}
