package redis

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "redis"

// init initializes package by register adapter
func init() {
	plugins.StorageRepo.Register(adapterName, redisAdapter{})
}

// redisAdapter represents adapter for redis storage
type redisAdapter struct {
}
