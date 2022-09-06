package facebook

import (
	"aureole/configs"
)

type (
	config struct {
		Filter       map[string]string `mapstructure:"filter" json:"filter"`
		ClientId     string            `mapstructure:"client_id" json:"client_id"`
		ClientSecret string            `mapstructure:"client_secret" json:"client_secret"`
		Scopes       []string          `mapstructure:"scopes" json:"scopes"`
		Fields       []string          `mapstructure:"fields" json:"fields"`
	}
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.Scopes, []string{"email"})
}
