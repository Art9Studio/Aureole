package google

import "aureole/internal/configs"

const (
	pathPrefix  = "/" + adapterName
	redirectUrl = "/login"
)

type config struct {
	ClientId     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	Scopes       []string `mapstructure:"scopes"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Scopes, []string{"https://www.googleapis.com/auth/userinfo.email"})
}
