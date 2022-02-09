package email

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "email"

// init initializes package by register adapter
func init() {
	plugins.AuthNRepo.Register(adapterName, adapter{})
}

// emailAdapter represents adapter for password based authentication
type adapter struct {
}

func (adapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &authn{rawConf: conf}
}
