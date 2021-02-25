package pwbased

import (
	authnTypes "aureole/plugins/authn/types"
)

type pwBased struct {
	ctx *Ctx
}

func (p pwBased) GetRoutes() []authnTypes.Route {
	return []authnTypes.Route{
		{
			Method:  "POST",
			Path:    p.ctx.PathPrefix,
			Handler: Auth(p.ctx),
		},
	}
}
