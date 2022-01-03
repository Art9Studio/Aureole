package email

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "email"

// init initializes package by register adapter
func init() {
	plugins.SenderRepo.Register(adapterName, emailAdapter{})
}

// emailAdapter represents adapter for the email messenger
type emailAdapter struct {
}
