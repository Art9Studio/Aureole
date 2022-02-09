package memory

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "memory"

// init initializes package by register adapter
func init() {
	plugins.StorageRepo.Register(adapterName, adapter{})
}

// adapter represents adapter for bigcache storage
type adapter struct {
}

func (adapter) Create(conf *configs.Storage) plugins.Storage {
	return &storage{rawConf: conf}
}
