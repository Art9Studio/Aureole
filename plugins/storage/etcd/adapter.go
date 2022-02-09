package etcd

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "etcd"

// init initializes package by register adapter
func init() {
	plugins.StorageRepo.Register(adapterName, adapter{})
}

// adapter represents adapter for etcd storage
type adapter struct {
}

func (adapter) Create(conf *configs.Storage) plugins.Storage {
	return &storage{rawConf: conf}
}
