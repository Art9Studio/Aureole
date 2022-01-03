package email

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "email"

// init initializes package by register adapter
func init() {
	plugins.AuthNRepo.Register(adapterName, emailAdapter{})
}

// emailAdapter represents adapter for password based authentication
type emailAdapter struct {
}
