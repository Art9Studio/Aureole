package argon2

import (
	"aureole/plugins/pwhasher"
)

// AdapterName is the internal name of the adapter
const AdapterName = "argon2"

// init initializes package by register adapter
func init() {
	pwhasher.RegisterAdapter(AdapterName, argon2Adapter{})
}

// argon2Adapter represents adapter for argon2 pwhasher algorithm
type argon2Adapter struct {
}
