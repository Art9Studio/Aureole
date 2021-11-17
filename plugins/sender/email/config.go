package email

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/sender/types"
)

type config struct {
	Host               string            `mapstructure:"host"`
	Username           string            `mapstructure:"username"`
	Password           string            `mapstructure:"password"`
	InsecureSkipVerify bool              `mapstructure:"insecure_skip_verify"`
	From               string            `mapstructure:"from"`
	Bcc                []string          `mapstructure:"bcc"`
	Cc                 []string          `mapstructure:"cc"`
	Templates          map[string]string `mapstructure:"templates"`
}

func (emailAdapter) Create(conf *configs.Sender) types.Sender {
	return &Email{rawConf: conf}
}
