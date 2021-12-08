package vault

import (
	"aureole/internal/plugins/storage"
)

// AdapterName is the internal name of the adapter
const AdapterName = "vault"

// init initializes package by register adapter
func init() {
	storage.Repository.Register(AdapterName, vaultAdapter{})
}

// vaultAdapter represents adapter for postgresql database
type vaultAdapter struct {
}
