package pwbased

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "password_based"

// init initializes package by registerConf adapter
func init() {
	plugins.AuthNRepo.Register(adapterName, adapter{})
}

// adapter represents adapter for password based authentication
type adapter struct {
}

func (adapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &authn{rawConf: conf}
}
