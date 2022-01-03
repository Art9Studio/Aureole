package apple

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "apple"

// init initializes package by register adapter
func init() {
	plugins.AuthNRepo.Register(adapterName, appleAdapter{})
}

type appleAdapter struct {
}
