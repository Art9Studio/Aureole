package email

import (
	"aureole/internal/plugins/authn"
)

// AdapterName is the internal name of the adapter
const AdapterName = "email"

// init initializes package by register adapter
func init() {
	authn.Repository.Register(AdapterName, emailAdapter{})
	authn.Repository.PluginApi.RegisterCollectionType(emailLinkCollType)
}

// emailAdapter represents adapter for password based authentication
type emailAdapter struct {
}
