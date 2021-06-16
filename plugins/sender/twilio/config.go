package twilio

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/sender/types"
)

type config struct {
	AccountSid string            `mapstructure:"account_sid"`
	AuthToken  string            `mapstructure:"auth_token"`
	From       string            `mapstructure:"from"`
	Templates  map[string]string `mapstructure:"templates"`
}

func (e twilioAdapter) Create(conf *configs.Sender) types.Sender {
	return &Twilio{rawConf: conf}
}
