package pwbased

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const (
	pathPrefix       = "/password-based"
	registerUrl      = "/register"
	resetUrl         = "/reset-password"
	resetConfirmUrl  = "/reset-password/confirm"
	verifyUrl        = "/verify-email"
	verifyConfirmUrl = "/verify-email/confirm"
)

type (
	config struct {
		MainHasher    string       `mapstructure:"main_hasher"`
		CompatHashers []string     `mapstructure:"compat_hashers"`
		Register      registerConf `mapstructure:"registerConf"`
		Reset         resetConf    `mapstructure:"password_reset"`
		Verif         verifConf    `mapstructure:"verification"`
	}

	registerConf struct {
		IsLoginAfter  bool `mapstructure:"login_after"`
		IsVerifyAfter bool `mapstructure:"verify_after"`
	}

	resetConf struct {
		Sender   string `mapstructure:"sender"`
		Template string `mapstructure:"template"`
		Exp      int    `mapstructure:"exp"`
	}

	verifConf struct {
		Sender   string `mapstructure:"sender"`
		Template string `mapstructure:"template"`
		Exp      int    `mapstructure:"exp"`
	}
)

func (pwBasedAdapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &pwBased{rawConf: conf}
}
