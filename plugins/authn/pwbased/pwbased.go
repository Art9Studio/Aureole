package pwbased

import (
	contextTypes "aureole/context/types"
	"aureole/internal/collections"
	"aureole/internal/plugins/authn"
	authnTypes "aureole/internal/plugins/authn/types"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/pwhasher/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"fmt"
	"path"
)

type pwBased struct {
	Conf           *сonfig
	ProjectContext *contextTypes.ProjectCtx
	AppName        string
	AuthzName      string
	PathPrefix     string
	PwHasher       types.PwHasher
	Storage        storageTypes.Storage
	IdentityColl   *collections.Collection
	Authorizer     authzTypes.Authorizer
}

func (p *pwBased) Initialize() error {
	projectCtx := authn.Repository.ProjectCtx

	hasher, ok := projectCtx.Hashers[p.Conf.MainHasher]
	if !ok {
		return fmt.Errorf("hasher named '%s' is not declared", p.Conf.MainHasher)
	}

	collection, ok := projectCtx.Collections[p.Conf.Collection]
	if !ok {
		return fmt.Errorf("collection named '%s' is not declared", p.Conf.Collection)
	}

	storage, ok := projectCtx.Storages[p.Conf.Storage]
	if !ok {
		return fmt.Errorf("storage named '%s' is not declared", p.Conf.Storage)
	}

	authorizer, ok := projectCtx.Apps[p.AppName].Authorizers[p.AuthzName]
	if !ok {
		return fmt.Errorf("authorizer named '%s' is not declared", p.AuthzName)
	}

	p.ProjectContext = projectCtx
	p.PwHasher = hasher
	p.IdentityColl = collection
	p.Storage = storage
	p.Authorizer = authorizer

	return p.Storage.CheckFeaturesAvailable([]string{p.IdentityColl.Type})
}

func (p *pwBased) GetRoutes() []authnTypes.Route {
	return []authnTypes.Route{
		{
			Method:  "POST",
			Path:    path.Clean(p.PathPrefix + p.Conf.Login.Path),
			Handler: Login(p),
		},
		{
			Method:  "POST",
			Path:    path.Clean(p.PathPrefix + p.Conf.Register.Path),
			Handler: Register(p),
		},
	}
}
