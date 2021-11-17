package memory

import (
	"aureole/internal/plugins/storage"
)

// AdapterName is the internal name of the adapter
const AdapterName = "memory"

// init initializes package by register adapter
func init() {
	storage.Repository.Register(AdapterName, memoryAdapter{})
}

// memoryAdapter represents adapter for bigcache storage
type memoryAdapter struct {
}
