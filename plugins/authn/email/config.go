package email

import (
	"aureole/internal/configs"
	authnTypes "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		Collection string        `mapstructure:"collection"`
		Storage    string        `mapstructure:"storage"`
		Login      login         `mapstructure:"login"`
		Register   register      `mapstructure:"register"`
		Link       magicLinkConf `mapstructure:"magic_link"`
	}

	login struct {
		Path      string            `mapstructure:"path"`
		FieldsMap map[string]string `mapstructure:"fields_map"`
	}

	register struct {
		Path      string            `mapstructure:"path"`
		FieldsMap map[string]string `mapstructure:"fields_map"`
	}

	magicLinkConf struct {
		Path       string `mapstructure:"path"`
		Collection string `mapstructure:"collection"`
		Sender     string `mapstructure:"sender"`
		Template   string `mapstructure:"template"`
		Token      token  `mapstructure:"token"`
	}

	token struct {
		Exp      int    `mapstructure:"exp"`
		HashFunc string `mapstructure:"hash_func"`
	}
)

func (e emailAdapter) Create(conf *configs.Authn) authnTypes.Authenticator {
	return &email{rawConf: conf}
}
