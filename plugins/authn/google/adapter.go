package google

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "google"

// init initializes package by register adapter
func init() {
	plugins.AuthNRepo.Register(adapterName, googleAdapter{})
}

type googleAdapter struct {
}
