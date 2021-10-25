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
	Url            *url.URL
	PathPrefix     string
	Identity       *identity.Identity
	Authenticators map[string]authnTypes.Authenticator
	Authorizers    map[string]authzTypes.Authorizer
}

func (a *App) GetName() string {
	return a.Name
}

func (a *App) GetUrl() (*url.URL, error) {
	if a.Url == nil {
		return nil, fmt.Errorf("can't find app url for app '%s'", a.Name)
	}
	return a.Url, nil
}

func (a *App) GetPathPrefix() string {
	return a.PathPrefix
}

func (a *App) GetIdentity() (*identity.Identity, error) {
	if a.Identity == nil {
		return nil, fmt.Errorf("can't find identity for app '%s'", a.Name)
	}
	return a.Identity, nil
}

func (a *App) GetAuthorizer(name string) (authzTypes.Authorizer, error) {
	authz, ok := a.Authorizers[name]
	if !ok || authz == nil {
		return nil, fmt.Errorf("can't find authorizer named '%s'", name)
	}
	return authz, nil
}
