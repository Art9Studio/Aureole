package apple

import (
	"aureole/configs"
)

type (
	config struct {
		Filter    map[string]string `mapstructure:"filter" json:"filter"`
		SecretKey string            `mapstructure:"secret_key" json:"secret_key"`
		PublicKey string            `mapstructure:"public_key" json:"public_key"`
		ClientId  string            `mapstructure:"client_id" json:"client_id"`
		TeamId    string            `mapstructure:"team_id" json:"team_id"`
		KeyId     string            `mapstructure:"key_id" json:"key_id"`
		Scopes    []string          `mapstructure:"scopes" json:"scopes"`
	}
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.Scopes, []string{"email"})
}
