package pwbased

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "password_based"

// init initializes package by registerConf adapter
func init() {
	plugins.AuthNRepo.Register(adapterName, pwBasedAdapter{})
}

// pwBasedAdapter represents adapter for password based authentication
type pwBasedAdapter struct {
}
