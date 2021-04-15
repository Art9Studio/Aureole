package pwbased

import (
	"aureole/configs"
	contextTypes "aureole/context/types"
	"aureole/internal/collections"
	"aureole/internal/plugins/authn"
	authnTypes "aureole/internal/plugins/authn/types"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/pwhasher/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"path"
)

type pwBased struct {
	rawConf        *configs.Authn
	conf           *сonfig
	projectContext *contextTypes.ProjectCtx
	pwHasher       types.PwHasher
	storage        storageTypes.Storage
	identityColl   *collections.Collection
	authorizer     authzTypes.Authorizer
}

func (p *pwBased) Initialize(appName string) error {
	projectCtx := authn.Repository.ProjectCtx
	adapterConf := &сonfig{}
	if err := mapstructure.Decode(p.rawConf.Config, adapterConf); err != nil {
		return err
	}
	adapterConf.setDefaults()

	p.conf = adapterConf
	p.projectContext = projectCtx

	hasher, ok := projectCtx.Hashers[p.conf.MainHasher]
	if !ok {
		return fmt.Errorf("hasher named '%s' is not declared", p.conf.MainHasher)
	}

	collection, ok := projectCtx.Collections[p.conf.Collection]
	if !ok {
		return fmt.Errorf("collection named '%s' is not declared", p.conf.Collection)
	}

	storage, ok := projectCtx.Storages[p.conf.Storage]
	if !ok {
		return fmt.Errorf("storage named '%s' is not declared", p.conf.Storage)
	}

	authorizer, ok := projectCtx.Apps[appName].Authorizers[p.rawConf.AuthzName]
	if !ok {
		return fmt.Errorf("authorizer named '%s' is not declared", p.rawConf.AuthzName)
	}

	p.pwHasher = hasher
	p.identityColl = collection
	p.storage = storage
	p.authorizer = authorizer

	return p.storage.CheckFeaturesAvailable([]string{p.identityColl.Type})
}

func (p *pwBased) GetRoutes() []authnTypes.Route {
	return []authnTypes.Route{
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
