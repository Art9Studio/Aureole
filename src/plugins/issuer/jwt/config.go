package jwt

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const refreshUrl = "/jwt/refresh"

type config struct {
	Iss           string     `mapstructure:"iss"`
	Sub           bool       `mapstructure:"sub"`
	Aud           []string   `mapstructure:"aud"`
	Nbf           int        `mapstructure:"nbf"`
	Iat           bool       `mapstructure:"iat"`
	AccessBearer  bearerType `mapstructure:"access_bearer"`
	RefreshBearer bearerType `mapstructure:"refresh_bearer"`
	SignKey       string     `mapstructure:"sign_key"`
	VerifyKeys    []string   `mapstructure:"verify_keys"`
	AccessExp     int        `mapstructure:"access_exp"`
	RefreshExp    int        `mapstructure:"refresh_exp"`
	TmplPath      string     `mapstructure:"payload"`
	NativeQueries string     `mapstructure:"native_queries"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Iss, "Aureole Server")
	configs.SetDefault(&c.Aud, []string{})
	configs.SetDefault(&c.Nbf, 0)
	configs.SetDefault(&c.Iat, true)
	configs.SetDefault(&c.Sub, true)
	configs.SetDefault(&c.AccessBearer, header)
	configs.SetDefault(&c.RefreshBearer, cookie)
	configs.SetDefault(&c.AccessExp, 900)
	configs.SetDefault(&c.RefreshExp, 7890000)
	configs.SetDefault(&c.VerifyKeys, []string{c.SignKey})
}

func (pluginCreator) Create(conf configs.PluginConfig) plugins.Plugin {
	return &jwtIssuer{rawConf: conf}
}
