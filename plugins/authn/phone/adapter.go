package phone

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "phone"

// init initializes package by register adapter
func init() {
	plugins.AuthNRepo.Register(adapterName, phoneAdapter{})
}

// phoneAdapter represents adapter for password based authentication
type phoneAdapter struct {
}
