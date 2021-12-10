package redis

import (
	"aureole/internal/plugins/storage"
)

// AdapterName is the internal name of the adapter
const AdapterName = "redis"

// init initializes package by register adapter
func init() {
	storage.Repository.Register(AdapterName, redisAdapter{})
}

// redisAdapter represents adapter for redis storage
type redisAdapter struct {
}
