package pwbased

import (
	contextTypes "aureole/context/types"
	"aureole/internal/collections"
	authnTypes "aureole/internal/plugins/authn/types"
	"aureole/internal/plugins/pwhasher/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"path"
)

type pwBased struct {
	Conf           *сonfig
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
