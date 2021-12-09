package jwt

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "jwt"

// init initializes package by register adapter
func init() {
	plugins.AuthZRepo.Register(adapterName, jwtAdapter{})
}

// jwtAdapter represents adapter for jwtAuthz authorization
type jwtAdapter struct {
}
