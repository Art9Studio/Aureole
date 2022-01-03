package memory

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "memory"

// init initializes package by register adapter
func init() {
	plugins.StorageRepo.Register(adapterName, memoryAdapter{})
}

// memoryAdapter represents adapter for bigcache storage
type memoryAdapter struct {
}
