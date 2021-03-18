package pwbased

import (
	"aureole/collections"
	contextTypes "aureole/context/types"
	authnTypes "aureole/plugins/authn/types"
	"aureole/plugins/pwhasher/types"
	storageTypes "aureole/plugins/storage/types"
	"path"
)

type pwBased struct {
	Conf           *сonf
	ProjectContext *contextTypes.ProjectCtx
	PathPrefix     string
	PwHasher       types.PwHasher
	Storage        storageTypes.Storage
	IdentityColl   *collections.Collection
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
