package file

import (
	"aureole/internal/plugins/storage"
)

// AdapterName is the internal name of the adapter
const AdapterName = "file"

// init initializes package by register adapter
func init() {
	storage.Repository.Register(AdapterName, fileAdapter{})
}

// fileAdapter represents adapter for postgresql database
type fileAdapter struct {
}
