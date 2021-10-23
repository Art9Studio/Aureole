package phone

import (
	"aureole/internal/plugins/authn"
)

// AdapterName is the internal name of the adapter
const AdapterName = "phone"

// init initializes package by register adapter
func init() {
	authn.Repository.Register(AdapterName, phoneAdapter{})
}

// phoneAdapter represents adapter for password based authentication
type phoneAdapter struct {
}
