package twilio

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

type config struct {
	AccountSid string            `mapstructure:"account_sid"`
	AuthToken  string            `mapstructure:"auth_token"`
	From       string            `mapstructure:"from"`
	Templates  map[string]string `mapstructure:"templates"`
}

func (twilioAdapter) Create(conf *configs.Sender) plugins.Sender {
	return &twilio{rawConf: conf}
}
