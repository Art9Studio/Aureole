package phone

import (
	"aureole/internal/configs"
	authnTypes "aureole/internal/plugins/authn/types"
)

type (
	config struct {
		Hasher           string       `mapstructure:"hasher"`
		Collection       string       `mapstructure:"collection"`
		VerificationColl string       `mapstructure:"verification_collection"`
		Storage          string       `mapstructure:"storage"`
		Sender           string       `mapstructure:"sender"`
		Template         string       `mapstructure:"template"`
		ResendUrl        string       `mapstructure:"resend_url"`
		Login            login        `mapstructure:"login"`
		Register         register     `mapstructure:"register"`
		Verification     verification `mapstructure:"verification"`
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

	verification struct {
		Path        string            `mapstructure:"path"`
		MaxAttempts int               `mapstructure:"max_attempts"`
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
