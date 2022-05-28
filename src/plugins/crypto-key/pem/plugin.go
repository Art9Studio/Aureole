package pem

import (
	"aureole/internal/core"
)

// name is the internal name of the plugin
const name = "pem"

// init initializes package by register plugin
func init() {
	core.Repo.Register(name, pemPlugin{})
}

// pemPlugin represents plugin for jwk
type pemPlugin struct {
}
