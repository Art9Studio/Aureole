package apple

import (
	"aureole/internal/configs"
	authnT "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		SecretKey   string   `mapstructure:"secret_key"`
		PublicKey   string   `mapstructure:"public_key"`
		ClientId    string   `mapstructure:"client_id"`
		TeamId      string   `mapstructure:"team_id"`
		KeyId       string   `mapstructure:"key_id"`
		Scopes      []string `mapstructure:"scopes"`
		PathPrefix  string
		RedirectUri string
	}
)

func (appleAdapter) Create(conf *configs.Authn) authnT.Authenticator {
	return &apple{rawConf: conf}
}
