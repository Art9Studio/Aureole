package yubikey

import (
	"aureole/internal/core"
)

// name is the internal name of the plugin
const name = "yubikey"

// init initializes package by register plugin
func init() {
	core.Repo.Register(name, yubikeyPlugin{})
}

type yubikeyPlugin struct {
}
