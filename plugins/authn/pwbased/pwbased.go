package pwbased

import (
	"aureole/configs"
	"aureole/internal/collections"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
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

func (p *pwBased) Initialize(appName string) error {
	pluginsApi := authn.Repository.PluginsApi
	adapterConf := &сonfig{}
	if err := mapstructure.Decode(p.rawConf.Config, adapterConf); err != nil {
		return err
	}
	adapterConf.setDefaults()

	p.conf = adapterConf

	hasher, err := pluginsApi.GetHasher(p.conf.MainHasher)
	if err != nil {
		return fmt.Errorf("hasher named '%s' is not declared", p.conf.MainHasher)
	}

	collection, err := pluginsApi.GetCollection(p.conf.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", p.conf.Collection)
	}

	storage, err := pluginsApi.GetStorage(p.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", p.conf.Storage)
	}

	authorizer, err := pluginsApi.GetAuthorizer(p.rawConf.AuthzName, appName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", p.rawConf.AuthzName)
	}

	p.pwHasher = hasher
	p.identityColl = collection
	p.storage = storage
	p.authorizer = authorizer

	return p.storage.CheckFeaturesAvailable([]string{p.identityColl.Type})
}

func (p *pwBased) GetRoutes() []*router.Route {
	return []*router.Route{
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
}
