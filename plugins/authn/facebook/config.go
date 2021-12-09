package facebook

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const (
	pathPrefix  = "/oauth2/facebook"
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

func (facebookAdapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &facebook{rawConf: conf}
}
