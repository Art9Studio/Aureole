package urls

import (
	"aureole/internal/plugins/admin"
)

// AdapterName is the internal name of the adapter
const AdapterName = "urls"

// init initializes package by register adapter
func init() {
	admin.Repository.Register(AdapterName, urlsAdapter{})
}

// urlsAdapter represents adapter for argon2 pwhasher algorithm
type urlsAdapter struct {
}
