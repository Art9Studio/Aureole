package jwk

import (
	"aureole/internal/plugins/cryptokey"
)

// AdapterName is the internal name of the adapter
const AdapterName = "jwk"

// init initializes package by register adapter
func init() {
	cryptokey.Repository.Register(AdapterName, jwkAdapter{})
}

// jwkAdapter represents adapter for jwk
type jwkAdapter struct {
}
