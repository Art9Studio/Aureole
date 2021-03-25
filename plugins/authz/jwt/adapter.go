package jwt

import (
	"aureole/internal/plugins/authz"
)

// AdapterName is the internal name of the adapter
const AdapterName = "jwt"

// init initializes package by register adapter
func init() {
	authz.Repository.Register(AdapterName, jwtAdapter{})
}

// jwtAdapter represents adapter for jwt authorization
type jwtAdapter struct {
}
