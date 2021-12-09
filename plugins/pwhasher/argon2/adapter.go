package argon2

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "argon2"

// init initializes package by register adapter
func init() {
	plugins.PWHasherRepo.Register(adapterName, argon2Adapter{})
}

// argon2Adapter represents adapter for argon2 pwhasher algorithm
type argon2Adapter struct {
}
