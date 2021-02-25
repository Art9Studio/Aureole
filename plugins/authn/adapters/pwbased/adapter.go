package pwbased

import (
	"aureole/plugins/authn"
)

// AdapterName is the internal name of the adapter
const AdapterName = "password_based"

// init initializes package by register adapter
func init() {
	authn.RegisterAdapter(AdapterName, pwBasedAdapter{})
}

// pwBasedAdapter represents adapter for password based authentication
type pwBasedAdapter struct {
}
