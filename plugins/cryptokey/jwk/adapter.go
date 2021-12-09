package jwk

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "jwk"

// init initializes package by register adapter
func init() {
	plugins.CryptoKeyRepo.Register(adapterName, jwkAdapter{})
}

// jwkAdapter represents adapter for jwk
type jwkAdapter struct {
}
