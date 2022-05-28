package jwt_webhook

import (
	"aureole/internal/plugin"
)

// name is the internal name of the plugin
const name = "jwt_webhook"

// init initializes package by register plugin
func init() {
	plugin.Repo.Register(name, plugin{})
}

// plugin represents plugin for password based authentication
type plugin struct {
}
