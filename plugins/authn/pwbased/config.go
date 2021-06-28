package pwbased

import (
	"aureole/internal/configs"
	authnTypes "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		MainHasher    string    `mapstructure:"main_hasher"`
		CompatHashers []string  `mapstructure:"compat_hashers"`
		Collection    string    `mapstructure:"collection"`
		Storage       string    `mapstructure:"storage"`
		Login         login     `mapstructure:"login"`
		Register      register  `mapstructure:"register"`
		Reset         resetConf `mapstructure:"password_reset"`
	}

	login struct {
		Path      string            `mapstructure:"path"`
		FieldsMap map[string]string `mapstructure:"fields_map"`
	}

	register struct {
		Path         string            `mapstructure:"path"`
		IsLoginAfter bool              `mapstructure:"login_after"`
		FieldsMap    map[string]string `mapstructure:"fields_map"`
	}

	resetConf struct {
		Path       string            `mapstructure:"path"`
		ConfirmUrl string            `mapstructure:"confirm_url"`
		Collection string            `mapstructure:"collection"`
		Sender     string            `mapstructure:"sender"`
		Template   string            `mapstructure:"template"`
		TokenExp   int               `mapstructure:"token_exp"`
		FieldsMap  map[string]string `mapstructure:"fields_map"`
	}
)

func (p pwBasedAdapter) Create(conf *configs.Authn) authnTypes.Authenticator {
	return &pwBased{rawConf: conf}
}
