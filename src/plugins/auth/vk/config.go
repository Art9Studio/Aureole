package vk

import (
	"aureole/internal/configs"
)

const (
	pathPrefix  = "/vk"
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
