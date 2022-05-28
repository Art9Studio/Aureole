package etcd

import (
	"aureole/internal/core"
)

// name is the internal name of the plugin
const name = "etcd"

// init initializes package by register plugin
func init() {
	core.Repo.Register(name, etcdPlugin{})
}

// etcdPlugin represents plugin for etcd storage
type etcdPlugin struct {
}
