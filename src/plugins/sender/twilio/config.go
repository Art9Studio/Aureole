package twilio

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	AccountSid string `mapstructure:"account_sid"`
	AuthToken  string `mapstructure:"auth_token"`
	From       string `mapstructure:"from"`
}

func (twilioPlugin) Create(conf configs.PluginConfig) plugins.Plugin {
	return &twilio{rawConf: conf}
}
