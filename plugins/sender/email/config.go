package email

import (
	"aureole/configs"
	"aureole/internal/plugins/sender/types"
	"github.com/mitchellh/mapstructure"
)

type config struct {
	Host      string            `mapstructure:"host"`
	Username  string            `mapstructure:"username"`
	Password  string            `mapstructure:"password"`
	From      string            `mapstructure:"from"`
	Bcc       []string          `mapstructure:"bcc"`
	Cc        []string          `mapstructure:"cc"`
	Templates map[string]string `mapstructure:"templates"`
}

func (e emailAdapter) Create(conf *configs.Sender) (types.Sender, error) {
	adapterConfMap := conf.Config
	adapterConf := &config{}

	err := mapstructure.Decode(adapterConfMap, adapterConf)
	if err != nil {
		return nil, err
	}

	return &Email{Conf: adapterConf}, nil
}
