package jwt

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "jwt"

// init initializes package by register adapter
func init() {
	plugins.AuthZRepo.Register(adapterName, adapter{})
}

// adapter represents adapter for authz authorization
type adapter struct {
}

func (adapter) Create(conf *configs.Authz) plugins.Authorizer {
	return &authz{rawConf: conf}
}
