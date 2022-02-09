package vk

import "aureole/internal/configs"

const (
	pathPrefix  = "/" + adapterName
	redirectUrl = "/login"
)

type config struct {
	ClientId     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	Scopes       []string `mapstructure:"scopes"`
	Fields       []string `mapstructure:"fields"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.Scopes, []string{"email"})
}
