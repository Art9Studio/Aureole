package standard

import "aureole/internal/plugin"

// name is the internal name of the plugin
const name = "standard"

// init initializes package by register plugin
func init() {
	plugin.Repo.Register(name, plugin{})
}

// plugin represents plugin for password based authentication
type plugin struct {
}
