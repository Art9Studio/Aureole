package phone

import (
	"aureole/internal/configs"
	authnTypes "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		Hasher       string    `mapstructure:"hasher"`
		Collection   string    `mapstructure:"collection"`
		Storage      string    `mapstructure:"storage"`
		Login        login     `mapstructure:"login"`
		Register     register  `mapstructure:"register"`
		Verification verifConf `mapstructure:"verification"`
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

	verifConf struct {
		Path        string            `mapstructure:"path"`
		ResendUrl   string            `mapstructure:"resend_url"`
		Collection  string            `mapstructure:"collection"`
		MaxAttempts int               `mapstructure:"max_attempts"`
		Sender      string            `mapstructure:"sender"`
		Template    string            `mapstructure:"template"`
		Code        verificationCode  `mapstructure:"code"`
		FieldsMap   map[string]string `mapstructure:"fields_map"`
	}

	verificationCode struct {
		Length   int    `mapstructure:"length"`
		Alphabet string `mapstructure:"alphabet"`
		Prefix   string `mapstructure:"prefix"`
		Exp      int    `mapstructure:"exp"`
	}
)

func (p phoneAdapter) Create(conf *configs.Authn) authnTypes.Authenticator {
	return &phone{rawConf: conf}
}
