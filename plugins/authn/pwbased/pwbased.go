package pwbased

import (
	"aureole/configs"
	"aureole/internal/collections"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/core"
	"aureole/internal/plugins/pwhasher/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"aureole/internal/router"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"path"
)

type pwBased struct {
	rawConf      *configs.Authn
	conf         *сonfig
	pwHasher     types.PwHasher
	storage      storageTypes.Storage
	identityColl *collections.Collection
	authorizer   authzTypes.Authorizer
}

func (p *pwBased) Initialize(appName string) (err error) {
	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	pluginsApi := authn.Repository.PluginsApi
	p.pwHasher, err = pluginsApi.GetHasher(p.conf.MainHasher)
	if err != nil {
		return fmt.Errorf("hasher named '%s' is not declared", p.conf.MainHasher)
	}

	p.identityColl, err = pluginsApi.GetCollection(p.conf.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", p.conf.Collection)
	}

	p.storage, err = pluginsApi.GetStorage(p.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", p.conf.Storage)
	}

	p.authorizer, err = pluginsApi.GetAuthorizer(p.rawConf.AuthzName, appName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", p.rawConf.AuthzName)
	}

	if err = p.storage.CheckFeaturesAvailable([]string{p.identityColl.Type}); err != nil {
		return err
	}

	createRoutes(pluginsApi, p)
	return err
}

func initConfig(rawConf *configs.RawConfig) (*сonfig, error) {
	adapterConf := &сonfig{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()

	return adapterConf, nil
}

func createRoutes(pluginsApi *core.PluginsApi, p *pwBased) {
	routes := []*router.Route{
		{
			Method:  "POST",
			Path:    path.Clean(p.rawConf.PathPrefix + p.conf.Login.Path),
			Handler: Login(p),
		},
		{
			Method:  "POST",
			Path:    path.Clean(p.rawConf.PathPrefix + p.conf.Register.Path),
			Handler: Register(p),
		},
	}
	pluginsApi.AddRoutes(routes)
}
