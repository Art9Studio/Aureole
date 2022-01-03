package urls

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "urls"

// init initializes package by register adapter
func init() {
	plugins.AdminRepo.Register(adapterName, urlsAdapter{})
}

// urlsAdapter represents adapter for argon2 pwhasher algorithm
type urlsAdapter struct {
}
