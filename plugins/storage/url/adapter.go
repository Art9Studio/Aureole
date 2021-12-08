package url

import (
	"aureole/internal/plugins/storage"
)

// AdapterName is the internal name of the adapter
const AdapterName = "url"

// init initializes package by register adapter
func init() {
	storage.Repository.Register(AdapterName, urlAdapter{})
}

// urlAdapter represents adapter for postgresql database
type urlAdapter struct {
}
