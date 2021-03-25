package session

import (
	"aureole/internal/plugins/authz"
)

// AdapterName is the internal name of the adapter
const AdapterName = "session"

// init initializes package by register adapter
func init() {
	authz.Repository.Register(AdapterName, sessionAdapter{})
}

// sessionAdapter represents adapter for session authorization
type sessionAdapter struct {
}
