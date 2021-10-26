package postgresql

import (
	"aureole/internal/plugins/storage"
)

// AdapterName is the internal name of the adapter
const AdapterName = "postgresql"

// init initializes package by register adapter
func init() {
	storage.Repository.Register(AdapterName, pgAdapter{})
}

// pgAdapter represents adapter for postgresql database
type pgAdapter struct {
}
