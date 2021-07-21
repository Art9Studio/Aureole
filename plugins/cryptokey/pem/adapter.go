package pem

import (
	"aureole/internal/plugins/cryptokey"
)

// AdapterName is the internal name of the adapter
const AdapterName = "pem"

// init initializes package by register adapter
func init() {
	cryptokey.Repository.Register(AdapterName, pemAdapter{})
}

// pemAdapter represents adapter for jwk
type pemAdapter struct {
}
