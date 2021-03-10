package argon2

import (
	"gouth/authN"
	"gouth/config"
)
import "github.com/mitchellh/mapstructure"

type Config struct {
	MainHasher    string   `mapstructure:"main_hasher"`
	CompatHashers []string `mapstructure:"compat_hashers"`
	Storage       string   `mapstructure:"storage"`
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
	return nil, nil
}
