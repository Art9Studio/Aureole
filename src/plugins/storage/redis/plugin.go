package redis

import (
	"aureole/internal/core"
)

// name is the internal name of the plugin
const name = "redis"

// init initializes package by register plugin
func init() {
	core.Repo.Register(name, redisPlugin{})
}

// redisPlugin represents plugin for redis storage
type redisPlugin struct {
}
