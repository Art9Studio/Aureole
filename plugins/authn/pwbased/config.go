package pwbased

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
	"aureole/plugins/authn/pwbased/pwhasher"
)

const (
	pathPrefix        = "/password-based"
	registerUrl       = "/register"
	resetUrl          = "/reset-password"
	resetConfirmUrl   = "/reset-password/confirm"
	verifyUrl         = "/verify-email"
	verifyConfirmUrl  = "/verify-email/confirm"
	defaultResetTmpl  = "Your password reset link: {{.link}}"
	defaultVerifyTmpl = "Click and verify you email: {{.link}}"
)

type config struct {
	MainHasher    pwhasher.Config   `mapstructure:"main_hasher"`
	CompatHashers []pwhasher.Config `mapstructure:"compat_hashers"`
	Register      struct {
		IsLoginAfter  bool `mapstructure:"login_after"`
		IsVerifyAfter bool `mapstructure:"verify_after"`
	} `mapstructure:"register"`
	Reset struct {
		Sender   string `mapstructure:"sender"`
		TmplPath string `mapstructure:"template"`
		Exp      int    `mapstructure:"exp"`
	} `mapstructure:"password_reset"`
	Verify struct {
		Sender   string `mapstructure:"sender"`
		TmplPath string `mapstructure:"template"`
		Exp      int    `mapstructure:"exp"`
	} `mapstructure:"verification"`
}

func (pwBasedAdapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &pwBased{rawConf: conf}
}
