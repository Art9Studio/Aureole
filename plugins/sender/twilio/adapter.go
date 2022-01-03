package twilio

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "twilio"

// init initializes package by register adapter
func init() {
	plugins.SenderRepo.Register(adapterName, twilioAdapter{})
}

// twilioAdapter represents adapter for the email messenger
type twilioAdapter struct {
}
