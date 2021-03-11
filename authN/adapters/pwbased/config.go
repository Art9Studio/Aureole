package pwbased

import (
	"github.com/mitchellh/mapstructure"
	"gouth/authN"
	"gouth/config"
)

type Config struct {
	// common info
	Path string
	// config
	MainHasher    string   `mapstructure:"main_hasher"`
	CompatHashers []string `mapstructure:"compat_hashers"`
	Collection    string   `mapstructure:"collection"`
	Identity      string   `mapstructure:"identity"`
	Password      string   `mapstructure:"password"`
}

func (p pwBasedAdapter) GetAuthNController(path string, configMap *config.RawConfig) (authN.Controller, error) {
	controllerConfig := Config{}
	controllerConfig.Path = path

	err := mapstructure.Decode(configMap, &controllerConfig)

	if err != nil {
		return nil, err
	}

	return &pwBased{&controllerConfig}, nil
}
