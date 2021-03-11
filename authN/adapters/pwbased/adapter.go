package pwbased

import (
	"gouth/authN"
)

// AdapterName is the internal name of the adapter
const AdapterName = "password_based"

// init initializes package by register adapter
func init() {
	authN.RegisterAdapter(AdapterName, pwBasedAdapter{})
}

// pwBasedAdapter represents adapter for password based authentication
type pwBasedAdapter struct {
}
