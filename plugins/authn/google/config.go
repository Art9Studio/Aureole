package google

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const (
	pathPrefix  = "/oauth2/google"
	redirectUrl = "/login"
)

type (
	config struct {
		ClientId     string   `mapstructure:"client_id"`
		ClientSecret string   `mapstructure:"client_secret"`
		Scopes       []string `mapstructure:"scopes"`
	}
)

func (googleAdapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &google{rawConf: conf}
}
