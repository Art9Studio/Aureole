package pwbased

import (
	"gouth/authN"
)

type pwBased struct {
	ctx *Ctx
}

func (p pwBased) GetRoutes() []authN.Route {
	return []authN.Route{
		{
			Method:  "POST",
			Path:    p.ctx.PathPrefix,
			Handler: Auth(p.ctx),
		},
	}
}
