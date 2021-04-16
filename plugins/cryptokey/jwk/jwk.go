package jwk

import (
	"aureole/configs"
	"context"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
	"net/url"
)

type Jwk struct {
	rawConf *configs.CryptoKey
	conf    *config
}

func (j *Jwk) Init() error {
	adapterConf := &config{}
	if err := mapstructure.Decode(j.rawConf.Config, adapterConf); err != nil {
		return err
	}
	j.conf = adapterConf

	return nil
}

func (j *Jwk) Get(path string) (jwk.Set, error) {
	if _, err := url.ParseRequestURI(path); err != nil {
		return jwk.ReadFile(path)
	}

	return jwk.Fetch(context.Background(), path)
}
