package facebook

import (
	"aureole/internal/configs"
)

const (
	pathPrefix  = "/facebook"
	redirectUrl = "/login"
)

type (
	config struct {
		Filter       map[string]string `mapstructure:"filter"`
		ClientId     string            `mapstructure:"client_id"`
		ClientSecret string            `mapstructure:"client_secret"`
		Scopes       []string          `mapstructure:"scopes"`
		Fields       []string          `mapstructure:"fields"`
	}
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.Scopes, []string{"email"})
}
