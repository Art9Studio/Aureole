package app

import (
	"aureole/internal/identity"
	fa2T "aureole/internal/plugins/2fa/types"
	auhnT "aureole/internal/plugins/authn/types"
	authzT "aureole/internal/plugins/authz/types"
	"fmt"
	"net/url"
	"regexp"
)

type App struct {
	Name            string
	Url             *url.URL
	PathPrefix      string
	IdentityManager identity.ManagerI
	Authenticators  map[string]auhnT.Authenticator
	Authorizer      authzT.Authorizer
	SecondFactor    fa2T.SecondFactor
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

func (a *App) GetSecondFactor() (fa2T.SecondFactor, error) {
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
