package twilio

import (
	"aureole/internal/configs"
	"aureole/internal/core"
)

type config struct {
	AccountSid string `mapstructure:"account_sid"`
	AuthToken  string `mapstructure:"auth_token"`
	From       string `mapstructure:"from"`
}

func (twilioPlugin) Create(conf configs.PluginConfig) core.Sender {
	return &twilio{rawConf: conf}
}
