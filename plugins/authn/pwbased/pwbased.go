package pwbased

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/pwhasher/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"aureole/internal/router/interface"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type pwBased struct {
	appName    string
	rawConf    *configs.Authn
	conf       *config
	identity   *identity.Identity
	pwHasher   types.PwHasher
	storage    storageTypes.Storage
	coll       *collections.Collection
	authorizer authzTypes.Authorizer
}

func (p *pwBased) Init(appName string) (err error) {
	p.appName = appName

	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi
	p.pwHasher, err = pluginApi.Project.GetHasher(p.conf.MainHasher)
	if err != nil {
		return fmt.Errorf("hasher named '%s' is not declared", p.conf.MainHasher)
	}

	p.coll, err = pluginApi.Project.GetCollection(p.conf.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", p.conf.Collection)
	}

	p.storage, err = pluginApi.Project.GetStorage(p.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", p.conf.Storage)
	}

	p.authorizer, err = pluginApi.Project.GetAuthorizer(p.rawConf.AuthzName, appName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", p.rawConf.AuthzName)
	}

	p.identity, err = pluginApi.Project.GetIdentity(appName)
	if err != nil {
		return fmt.Errorf("identity in app '%s' is not declared", appName)
	}

	if err = p.storage.CheckFeaturesAvailable([]string{p.coll.Type}); err != nil {
		return err
	}

	createRoutes(p)
	return nil
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()

	return adapterConf, nil
}

func createRoutes(p *pwBased) {
	routes := []*_interface.Route{
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.Login.Path,
			Handler: Login(p),
		},
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.Register.Path,
			Handler: Register(p),
		},
	}
	authn.Repository.PluginApi.Router.Add(p.appName, routes)
}
