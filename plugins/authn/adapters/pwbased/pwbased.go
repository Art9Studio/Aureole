package pwbased

import (
	"aureole/collections"
	contextTypes "aureole/context/types"
	authnTypes "aureole/plugins/authn/types"
	"aureole/plugins/pwhasher/types"
	types2 "aureole/plugins/storage/types"
)

type pwBased struct {
	Conf           *Conf
	ProjectContext *contextTypes.ProjectCtx
	PathPrefix     string
	PwHasher       types.PwHasher
	Storage        types2.Storage
	IdentityColl   *collections.Collection
	Identity       string
	Password       string
}

func (p *pwBased) GetRoutes() []authnTypes.Route {
	return []authnTypes.Route{
		{
			Method:  "POST",
			Path:    p.PathPrefix,
			Handler: Auth(p),
		},
	}
}
