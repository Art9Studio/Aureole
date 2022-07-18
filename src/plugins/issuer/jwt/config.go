package jwt

import (
	"aureole/internal/configs"
)

const refreshUrl = "/jwt/refresh"

type config struct {
	Iss                string    `mapstructure:"iss" json:"iss"`
	Sub                bool      `mapstructure:"sub" json:"sub"`
	Aud                []string  `mapstructure:"aud" json:"aud"`
	Nbf                int       `mapstructure:"nbf" json:"nbf"`
	Iat                bool      `mapstructure:"iat" json:"iat"`
	AccessTokenBearer  tokenResp `mapstructure:"access_bearer" json:"access_bearer"`
	RefreshTokenBearer tokenResp `mapstructure:"refresh_bearer" json:"refresh_bearer"`
	SignKey            string    `mapstructure:"sign_key" json:"sign_key"`
	VerifyKeys         []string  `mapstructure:"verify_keys" json:"verify_keys"`
	AccessExp          int       `mapstructure:"access_exp" json:"access_exp"`
	RefreshExp         int       `mapstructure:"refresh_exp" json:"refresh_exp"`
	TmplPath           string    `mapstructure:"payload" json:"payload"`
	NativeQueries      string    `mapstructure:"native_queries" json:"native_queries"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Iss, "Aureole Server")
	configs.SetDefault(&c.Aud, []string{})
	configs.SetDefault(&c.Nbf, 0)
	configs.SetDefault(&c.Iat, true)
	configs.SetDefault(&c.Sub, true)
	configs.SetDefault(&c.AccessTokenBearer, body)
	configs.SetDefault(&c.RefreshTokenBearer, body)
	configs.SetDefault(&c.AccessExp, 900)
	configs.SetDefault(&c.RefreshExp, 7890000)
	configs.SetDefault(&c.VerifyKeys, []string{c.SignKey})
}
