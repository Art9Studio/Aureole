package facebook

import (
	"aureole/internal/plugins/authn"
)

// AdapterName is the internal name of the adapter
const AdapterName = "facebook"

// init initializes package by register adapter
func init() {
	authn.Repository.Register(AdapterName, facebookAdapter{})
}

type facebookAdapter struct {
}
