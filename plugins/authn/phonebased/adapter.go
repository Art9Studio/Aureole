package phonebased

import (
	"aureole/internal/plugins/authn"
)

// AdapterName is the internal name of the adapter
const AdapterName = "phone_based"

// init initializes package by register adapter
func init() {
	authn.Repository.Register(AdapterName, phoneBasedAdapter{})
	authn.Repository.PluginApi.RegisterCollectionType(phoneVerifyCollType)
}

// phoneBasedAdapter represents adapter for password based authentication
type phoneBasedAdapter struct {
}
