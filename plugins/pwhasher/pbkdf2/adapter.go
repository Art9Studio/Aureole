package pbkdf2

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "pbkdf2"

// init initializes package by register adapter
func init() {
	plugins.PWHasherRepo.Register(adapterName, pbkdf2Adapter{})
}

// pbkdf2Adapter represents adapter for pbkdf2 pwhasher algorithm
type pbkdf2Adapter struct {
}
