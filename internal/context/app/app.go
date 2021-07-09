package app

import (
	"aureole/internal/identity"
	authnTypes "aureole/internal/plugins/authn/types"
	authzTypes "aureole/internal/plugins/authz/types"
)

type App struct {
	Host           string
	PathPrefix     string
	Identity       *identity.Identity
	Authenticators []authnTypes.Authenticator
	Authorizers    map[string]authzTypes.Authorizer
}
