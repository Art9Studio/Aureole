package jwk

import (
	"aureole/configs"
	"aureole/internal/plugins/cryptokey/types"
	"github.com/mitchellh/mapstructure"
)

type config struct {
	Path string `mapstructure:"path"`
}

func (j jwkAdapter) Create(conf *configs.CryptoKey) (types.CryptoKey, error) {
	adapterConfMap := conf.Config
	adapterConf := &config{}

	err := mapstructure.Decode(adapterConfMap, adapterConf)
	if err != nil {
		return nil, err
	}

	return initAdapter(conf, adapterConf)
}

func initAdapter(conf *configs.CryptoKey, adapterConf *config) (*Jwk, error) {
	return &Jwk{Conf: adapterConf}, nil
}
