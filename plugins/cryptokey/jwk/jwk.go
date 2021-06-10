package jwk

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/cryptokey"
	_interface "aureole/internal/router/interface"
	"context"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/mitchellh/mapstructure"
)

type Jwk struct {
	rawConf    *configs.CryptoKey
	conf       *config
	privateSet jwk.Set
	publicSet  jwk.Set
}

func (j *Jwk) Init() (err error) {
	if j.conf, err = initConfig(&j.rawConf.Config); err != nil {
		return err
	}

	if j.privateSet, err = createPrivateSet(j.conf.Path); err != nil {
		return err
	}
	if j.publicSet, err = jwk.PublicSetOf(j.privateSet); err != nil {
		return err
	}

	createRoutes(j)
	return nil
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}

	return adapterConf, nil
}

func createPrivateSet(path string) (privateSet jwk.Set, err error) {
	privateSet, err = jwk.Fetch(context.Background(), path)
	if err != nil {
		privateSet, err = jwk.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}

	return privateSet, nil
}

func createRoutes(j *Jwk) {
	routes := []*_interface.Route{
		{
			Method:  "GET",
			Path:    j.rawConf.PathPrefix + "/jwk",
			Handler: GetJwkKeys(j),
		},
		{
			Method:  "GET",
			Path:    j.rawConf.PathPrefix + "/pem",
			Handler: GetPemKeys(j),
		},
	}
	cryptokey.Repository.PluginApi.Router.AddProjectRoutes(routes)
}

func (j *Jwk) GetPrivateSet() jwk.Set {
	return j.privateSet
}

func (j *Jwk) GetPublicSet() jwk.Set {
	return j.publicSet
}
