package twilio

import (
	"aureole/internal/core"
)

// name is the internal name of the plugin
const name = "twilio"

// init initializes package by register plugin
func init() {
	core.Repo.Register(name, twilioPlugin{})
}

// twilioPlugin represents plugin for the email messenger
type twilioPlugin struct {
}
