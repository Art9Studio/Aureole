package authenticator

import (
	factor2 "aureole/internal/plugins/2fa"
)

// AdapterName is the internal name of the adapter
const AdapterName = "google_authenticator"

// init initializes package by register adapter
func init() {
	factor2.Repository.Register(AdapterName, gauthAdapter{})
}

type gauthAdapter struct {
}
