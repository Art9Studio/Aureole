package google

import "aureole/internal/configs"

const (
	pathPrefix  = "/google"
	redirectUrl = "/login"
)

type (
	config struct {
		Filter       map[string]string `mapstructure:"filter" json:"filter"`
		ClientId     string            `mapstructure:"client_id" json:"client_id"`
		ClientSecret string            `mapstructure:"client_secret" json:"client_secret"`
		Scopes       []string          `mapstructure:"scopes" json:"scopes"`
	}
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.Scopes, []string{"https://www.googleapis.com/auth/userinfo.email"})
}
