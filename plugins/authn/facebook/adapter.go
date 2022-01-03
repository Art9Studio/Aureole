package facebook

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "facebook"

// init initializes package by register adapter
func init() {
	plugins.AuthNRepo.Register(adapterName, facebookAdapter{})
}

type facebookAdapter struct {
}
