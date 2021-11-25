package app

import (
	"aureole/internal/identity"
	mfaT "aureole/internal/plugins/2fa/types"
	authnT "aureole/internal/plugins/authn/types"
	authzT "aureole/internal/plugins/authz/types"
	"fmt"
	"net/url"
	"regexp"
)

type App struct {
	Name            string
	Url             *url.URL
	PathPrefix      string
	AuthSessionExp  int
	IdentityManager identity.ManagerI
	Authenticators  map[string]authnT.Authenticator
	Authorizer      authzT.Authorizer
	SecondFactor    mfaT.SecondFactor
}

func (a *App) GetName() string {
	return a.Name
}

func (a *App) GetUrl() (url.URL, error) {
	if a.Url == nil {
		return url.URL{}, fmt.Errorf("can't find app url for app '%s'", a.Name)
	}
	return *a.Url, nil
}

func (a *App) GetPathPrefix() string {
	return a.PathPrefix
}

func (a *App) GetAuthSessionExp() int {
	return a.AuthSessionExp
}

func (a *App) GetIdentityManager() (identity.ManagerI, error) {
	if a.IdentityManager == nil {
		return nil, fmt.Errorf("can't find identity for app '%s'", a.Name)
	}
	return a.IdentityManager, nil
}

func (a *App) GetAuthorizer() (authzT.Authorizer, error) {
	if a.Authorizer == nil {
		return nil, fmt.Errorf("can't find authorizer for app '%s'", a.Name)
	}
	return a.Authorizer, nil
}

func (a *App) GetSecondFactor() (mfaT.SecondFactor, error) {
	if a.SecondFactor == nil {
		return nil, fmt.Errorf("can't find second factor for app '%s'", a.Name)
	}
	return a.SecondFactor, nil
}

func (*App) Filter(fields, filters map[string]string) (bool, error) {
	for fieldName, pattern := range filters {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false, err
		}
		if !re.MatchString(fields[fieldName]) {
			return false, nil
		}
	}
	return true, nil
}
