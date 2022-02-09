package phone

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "phone"

// init initializes package by register adapter
func init() {
	plugins.AuthNRepo.Register(adapterName, adapter{})
}

// adapter represents adapter for password based authentication
type adapter struct {
}

func (adapter) Create(conf *configs.Authn) plugins.Authenticator {
	return &authn{rawConf: conf}
}
