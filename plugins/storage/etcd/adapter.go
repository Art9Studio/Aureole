package etcd

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "etcd"

// init initializes package by register adapter
func init() {
	plugins.StorageRepo.Register(adapterName, etcdAdapter{})
}

// etcdAdapter represents adapter for etcd storage
type etcdAdapter struct {
}
