package pwbased

import (
	"github.com/mitchellh/mapstructure"
	"gouth/authN"
	"gouth/config"
)

type Config struct {
	MainHasher    string   `mapstructure:"main_hasher"`
	CompatHashers []string `mapstructure:"compat_hashers"`
	Collection    string   `mapstructure:"collection"`
	UserUnique    string   `mapstructure:"user_unique"`
	UserConfirm   string   `mapstructure:"user_confirm"`
}

func (p pwBasedAdapter) GetAuthNController(configMap *config.RawConfig) (authN.Controller, error) {
	controllerConfig := Config{}
	err := mapstructure.Decode(configMap, &controllerConfig)

	if err != nil {
		return nil, err
	}

	return &pwBased{&controllerConfig}, nil
}
