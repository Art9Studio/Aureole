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
		Verif         verifConf `mapstructure:"verification"`
	}

	login struct {
		Path string
	}

	register struct {
		Path          string
		IsLoginAfter  bool `mapstructure:"login_after"`
		IsVerifyAfter bool `mapstructure:"verify_after"`
	}

	resetConf struct {
		Path       string
		ConfirmUrl string
		Sender     string `mapstructure:"sender"`
		Template   string `mapstructure:"template"`
		Exp        int    `mapstructure:"exp"`
	}

	verifConf struct {
		Path       string
		ConfirmUrl string
		Sender     string `mapstructure:"sender"`
		Template   string `mapstructure:"template"`
		Exp        int    `mapstructure:"exp"`
	}
)

func (p pwBasedAdapter) Create(conf *configs.Authn) authnTypes.Authenticator {
	return &pwBased{rawConf: conf}
}
