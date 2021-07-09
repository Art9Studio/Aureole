package pwbased

import (
	"aureole/internal/plugins/authn"
)

// AdapterName is the internal name of the adapter
const AdapterName = "password_based"

// init initializes package by register adapter
func init() {
	authn.Repository.Register(AdapterName, pwBasedAdapter{})
	authn.Repository.PluginApi.RegisterCollectionType(passwordColType)
	authn.Repository.PluginApi.RegisterCollectionType(resetColType)
}

// pwBasedAdapter represents adapter for password based authentication
type pwBasedAdapter struct {
}
