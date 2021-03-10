package pwbased

import (
	"gouth/authN"
)

// AdapterName is the internal name of the adapter
const AdapterName = "pwbased"

// init initializes package by register adapter
func init() {
	authN.RegisterAdapter(AdapterName, pwBasedAdapter{})
}

// pwBasedAdapter represents adapter for password based authentication
type pwBasedAdapter struct {
}
