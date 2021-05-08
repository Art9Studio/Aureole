package jwt

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/authz/types"
	"github.com/lestrrat-go/jwx/jwa"
)

type config struct {
	Iss           string                 `mapstructure:"iss"`
	Sub           bool                   `mapstructure:"sub"`
	Aud           []string               `mapstructure:"aud"`
	Nbf           int                    `mapstructure:"nbf"`
	Iat           bool                   `mapstructure:"iat"`
	Jti           int                    `mapstructure:"jti"`
	AccessBearer  bearerType             `mapstructure:"access_bearer"`
	RefreshBearer bearerType             `mapstructure:"refresh_bearer"`
	Alg           jwa.SignatureAlgorithm `mapstructure:"alg"`
	SignKey       string                 `mapstructure:"sign_key"`
	VerifyKeys    []string               `mapstructure:"verify_keys"`
	AccessExp     int                    `mapstructure:"access_exp"`
	RefreshExp    int                    `mapstructure:"refresh_exp"`
	RefreshUrl    string                 `mapstructure:"refresh_url"`
	Payload       string                 `mapstructure:"payload"`
}

func (j jwtAdapter) Create(conf *configs.Authz) types.Authorizer {
	return &jwtAuthz{rawConf: conf}
}
