package etcd

import (
	"aureole/internal/plugins/storage"
)

// AdapterName is the internal name of the adapter
const AdapterName = "etcd"

// init initializes package by register adapter
func init() {
	storage.Repository.Register(AdapterName, etcdAdapter{})
}

// etcdAdapter represents adapter for etcd storage
type etcdAdapter struct {
}
