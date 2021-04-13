package session

import (
	"aureole/configs"
	"aureole/internal/plugins/authz/types"
	"github.com/mitchellh/mapstructure"
)

type config struct {
	Collection    string `mapstructure:"collection"`
	Storage       string `mapstructure:"storage"`
	Domain        string `mapstructure:"domain"`
	Path          string `mapstructure:"path"`
	MaxAge        int    `mapstructure:"max_age"`
	Secure        bool   `mapstructure:"secure"`
	HttpOnly      bool   `mapstructure:"http_only"`
	SameSite      string `mapstructure:"same_site"`
	CleanInterval int    `mapstructure:"clean_interval"`
}

func (s sessionAdapter) Create(conf *configs.Authz) (types.Authorizer, error) {
	adapterConfMap := conf.Config
	adapterConf := &config{}

	err := mapstructure.Decode(adapterConfMap, adapterConf)
	if err != nil {
		return nil, err
	}

	adapterConf.setDefaults()

	return &session{Conf: adapterConf}, nil
}
