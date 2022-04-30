package pem

import (
	"aureole/internal/plugins"
)

// name is the internal name of the plugin
const name = "pem"

// init initializes package by register plugin
func init() {
	plugins.Repo.Register(name, pemPlugin{})
}

// pemPlugin represents plugin for jwk
type pemPlugin struct {
}
