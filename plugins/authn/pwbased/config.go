package pwbased

import (
	"aureole/configs"
	"aureole/internal/plugins/authn/types"
	"github.com/mitchellh/mapstructure"
)

type (
	сonfig struct {
		MainHasher    string   `mapstructure:"main_hasher"`
		CompatHashers []string `mapstructure:"compat_hashers"`
		Collection    string   `mapstructure:"collection"`
		Storage       string   `mapstructure:"storage"`
		Login         login    `mapstructure:"login"`
		Register      register `mapstructure:"register"`
	}

	login struct {
		Path      string            `mapstructure:"path"`
		FieldsMap map[string]string `mapstructure:"fields_map"`
	}

	register struct {
		Path         string            `mapstructure:"path"`
		IsLoginAfter bool              `mapstructure:"login_after"`
		FieldsMap    map[string]string `mapstructure:"fields_map"`
	}
)

func (p pwBasedAdapter) Create(appName string, conf *configs.Authn) (types.Authenticator, error) {
	adapterConfMap := conf.Config
	adapterConf := &сonfig{}

	err := mapstructure.Decode(adapterConfMap, adapterConf)
	if err != nil {
		return nil, err
	}

	adapterConf.setDefaults()

	return &pwBased{
		Conf:       adapterConf,
		AppName:    appName,
		AuthzName:  conf.AuthzName,
		PathPrefix: conf.PathPrefix,
	}, nil
}
