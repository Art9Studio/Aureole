package vk

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const (
	pathPrefix  = "/oauth2/vk"
	redirectUrl = "/login"
)

type (
	config struct {
		ClientId     string   `mapstructure:"client_id"`
		ClientSecret string   `mapstructure:"client_secret"`
		Scopes       []string `mapstructure:"scopes"`
		Fields       []string `mapstructure:"fields"`
	}
)

func (vkAdapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &vk{rawConf: conf}
}
