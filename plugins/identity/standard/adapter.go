package standard

import "aureole/internal/plugins"

// adapterName is the internal name of the adapter
const adapterName = "standard"

// init initializes package by register adapter
func init() {
	plugins.IDManagerRepo.Register(adapterName, adapter{})
}

// adapter represents adapter for password based authentication
type adapter struct {
}
