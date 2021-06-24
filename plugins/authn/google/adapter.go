package google

import (
	"aureole/internal/plugins/authn"
)

// AdapterName is the internal name of the adapter
const AdapterName = "google"

// init initializes package by register adapter
func init() {
	authn.Repository.Register(AdapterName, googleAdapter{})
	authn.Repository.PluginApi.RegisterCollectionType(oauthCollType)
}

type googleAdapter struct {
}
