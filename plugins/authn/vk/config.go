package vk

import (
	"aureole/internal/configs"
	authnTypes "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		Coll         string   `mapstructure:"collection"`
		Storage      string   `mapstructure:"storage"`
		ClientId     string   `mapstructure:"client_id"`
		ClientSecret string   `mapstructure:"client_secret"`
		Scopes       []string `mapstructure:"scopes"`
		RedirectUri  string   `mapstructure:"redirect_uri"`
		Fields       []string `mapstructure:"fields"`
	}
)

func (v vkAdapter) Create(conf *configs.Authn) authnTypes.Authenticator {
	return &vk{rawConf: conf}
}
