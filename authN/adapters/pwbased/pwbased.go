package pwbased

import (
	"gouth/authN"
)

type pwBased struct {
	conf *Config
}

func (p pwBased) GetRoutes() []authN.Route {
	panic("implement me")
}
