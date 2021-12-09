package pem

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "pem"

// init initializes package by register adapter
func init() {
	plugins.CryptoKeyRepo.Register(adapterName, pemAdapter{})
}

// pemAdapter represents adapter for jwk
type pemAdapter struct {
}
