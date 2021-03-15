package postgresql

import (
	"aureole/plugins/storage"
)

// AdapterName is the internal name of the adapter
const AdapterName = "postgresql"

var AdapterFeatures = map[string]bool{"identity": true, "sessions": true}

// init initializes package by register adapter
func init() {
	storage.RegisterAdapter(AdapterName, pgAdapter{})
}

// pgAdapter represents adapter for postgresql database
type pgAdapter struct {
}
