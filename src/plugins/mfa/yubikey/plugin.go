package yubikey

import (
	"aureole/internal/plugins"
)

// name is the internal name of the plugin
const name = "yubikey"

// init initializes package by register plugin
func init() {
	plugins.Repo.Register(name, yubikeyPlugin{})
}

type yubikeyPlugin struct {
}
