package email

import (
	"aureole/internal/plugins/sender"
)

// AdapterName is the internal name of the adapter
const AdapterName = "email"

// init initializes package by register adapter
func init() {
	sender.Repository.Register(AdapterName, emailAdapter{})
}

// emailAdapter represents adapter for the email messenger
type emailAdapter struct {
}
