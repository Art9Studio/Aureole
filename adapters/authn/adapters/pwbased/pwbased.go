package pwbased

import (
	"gouth/adapters/authn"
)

type pwBased struct {
	ctx *Ctx
}

func (p pwBased) GetRoutes() []authn.Route {
	return []authn.Route{
		{
			Method:  "POST",
			Path:    p.ctx.PathPrefix,
			Handler: Auth(p.ctx),
		},
	}
}
