package authenticator

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "google_authenticator"

// init initializes package by register adapter
func init() {
	plugins.SecondFactorRepo.Register(adapterName, gauthAdapter{})
}

type gauthAdapter struct {
}
