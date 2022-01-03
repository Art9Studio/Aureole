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

func (jwtAdapter) Create(conf *configs.Authz) plugins.Authorizer {
	return &jwtAuthz{rawConf: conf}
}
