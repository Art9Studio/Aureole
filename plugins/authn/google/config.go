package google

import (
	"aureole/internal/configs"
	authnT "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		ClientId     string   `mapstructure:"client_id"`
		ClientSecret string   `mapstructure:"client_secret"`
		Scopes       []string `mapstructure:"scopes"`
		PathPrefix   string
		RedirectUri  string
	}
)

func (googleAdapter) Create(conf *configs.Authn) authnT.Authenticator {
	return &google{rawConf: conf}
}
