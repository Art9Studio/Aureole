package pwbased

import (
	"gouth/authN"
)

type pwBased struct {
	conf *Config
}

func (p pwBased) GetRoutes() []authN.Route {
	return []authN.Route{
		{
			Method:  "POST",
			Path:    p.conf.Path,
			Handler: Auth,
		},
	}
}
