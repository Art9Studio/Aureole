package apple

import (
	"aureole/internal/configs"
)

const (
	pathPrefix  = "/apple"
	redirectUrl = "/login"
)

type (
	config struct {
		Filter    map[string]string `mapstructure:"filter"`
		SecretKey string            `mapstructure:"secret_key"`
		PublicKey string            `mapstructure:"public_key"`
		ClientId  string            `mapstructure:"client_id"`
		TeamId    string            `mapstructure:"team_id"`
		KeyId     string            `mapstructure:"key_id"`
		Scopes    []string          `mapstructure:"scopes"`
	}
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.Scopes, []string{"email"})
}
