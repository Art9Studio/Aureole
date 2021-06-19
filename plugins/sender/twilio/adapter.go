package twilio

import (
	"aureole/internal/plugins/sender"
)

// AdapterName is the internal name of the adapter
const AdapterName = "twilio"

// init initializes package by register adapter
func init() {
	sender.Repository.Register(AdapterName, twilioAdapter{})
}

// twilioAdapter represents adapter for the email messenger
type twilioAdapter struct {
}
