package pbkdf2

import (
	"aureole/plugins/pwhasher"
)

// AdapterName is the internal name of the adapter
const AdapterName = "pbkdf2"

// init initializes package by register adapter
func init() {
	pwhasher.Repository.Register(AdapterName, pbkdf2Adapter{})
}

// pbkdf2Adapter represents adapter for pbkdf2 pwhasher algorithm
type pbkdf2Adapter struct {
}
