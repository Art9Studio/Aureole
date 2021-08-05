package apple

import (
	"aureole/internal/plugins/authn"
)

// AdapterName is the internal name of the adapter
const AdapterName = "apple"

// init initializes package by register adapter
func init() {
	authn.Repository.Register(AdapterName, appleAdapter{})
}

type appleAdapter struct {
}
