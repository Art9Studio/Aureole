package email

import (
	"aureole/internal/plugins"
)

// name is the internal name of the plugin
const name = "email"

// init initializes package by register plugin
func init() {
	plugins.Repo.Register(name, emailPlugin{})
}

// emailPlugin represents plugin for the email messenger
type emailPlugin struct {
}
