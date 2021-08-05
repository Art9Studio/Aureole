package app

import (
	"aureole/internal/identity"
	authnTypes "aureole/internal/plugins/authn/types"
	authzTypes "aureole/internal/plugins/authz/types"
	"fmt"
	"net/url"
)

type App struct {
	Name           string
	Url            url.URL
	PathPrefix     string
	Identity       *identity.Identity
	Authenticators []authnTypes.Authenticator
	Authorizers    map[string]authzTypes.Authorizer
}

func (a *App) GetName() string {
	return a.Name
}

func (a *App) GetUrl() url.URL {
	return a.Url
}

func (a *App) GetPathPrefix() string {
	return a.PathPrefix
}

func (a *App) GetIdentity() *identity.Identity {
	return a.Identity
}

func (a *App) GetAuthorizer(name string) (authzTypes.Authorizer, error) {
	authz, ok := a.Authorizers[name]
	if !ok {
		return nil, fmt.Errorf("can't find authorizer named '%s'", name)
	}
	return authz, nil
}
