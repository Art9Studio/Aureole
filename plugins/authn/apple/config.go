package apple

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const (
	pathPrefix  = "/" + adapterName
	redirectUrl = "/login"
)

type (
	config struct {
		SecretKey string   `mapstructure:"secret_key"`
		PublicKey string   `mapstructure:"public_key"`
		ClientId  string   `mapstructure:"client_id"`
		TeamId    string   `mapstructure:"team_id"`
		KeyId     string   `mapstructure:"key_id"`
		Scopes    []string `mapstructure:"scopes"`
	}
)

func (appleAdapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &apple{rawConf: conf}
}
