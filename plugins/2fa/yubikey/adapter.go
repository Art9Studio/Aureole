package yubikey

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "yubikey"

// init initializes package by register adapter
func init() {
	plugins.SecondFactorRepo.Register(adapterName, yubikeyAdapter{})
}

type yubikeyAdapter struct {
}
