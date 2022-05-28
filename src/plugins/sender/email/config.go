package email

import (
	"aureole/internal/configs"
	"aureole/internal/core"
)

type config struct {
	Host               string   `mapstructure:"host"`
	Username           string   `mapstructure:"username"`
	Password           string   `mapstructure:"password"`
	InsecureSkipVerify bool     `mapstructure:"insecure_skip_verify"`
	From               string   `mapstructure:"from"`
	Bcc                []string `mapstructure:"bcc"`
	Cc                 []string `mapstructure:"cc"`
}

func (emailPlugin) Create(conf configs.PluginConfig) core.Sender {
	return &email{rawConf: conf}
}
