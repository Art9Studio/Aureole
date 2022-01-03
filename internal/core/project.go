package core

import (
	"aureole/internal/identity"
	"aureole/internal/plugins"
	"errors"
	"fmt"
	"net/url"
	"regexp"
)

type (
	Project struct {
		apiVersion    string
		testRun       bool
		pingPath      string
		service       service
		apps          map[string]*App
		authorizers   map[string]plugins.Authorizer
		secondFactors map[string]plugins.SecondFactor
		storages      map[string]plugins.Storage
		keyStorages   map[string]plugins.KeyStorage
		hashers       map[string]plugins.PWHasher
		senders       map[string]plugins.Sender
		cryptoKeys    map[string]plugins.CryptoKey
		admins        map[string]plugins.Admin
	}

	service struct {
		signKey plugins.CryptoKey
		encKey  plugins.CryptoKey
		storage plugins.Storage
	}
)

func (p *Project) GetAPIVersion() string {
	return p.apiVersion
}

func (p *Project) IsTestRun() bool {
	return p.testRun
}

func (p *Project) GetPingPath() string {
	return p.pingPath
}

func (p *Project) GetApp(name string) (*App, error) {
	a, ok := p.apps[name]
	if !ok {
		return nil, fmt.Errorf("can't find app named '%s'", name)
	}

	return a, nil
}

func (p *Project) GetAuthorizer(name string) (plugins.Authorizer, error) {
	a, ok := p.authorizers[name]
	if !ok || a == nil {
		return nil, fmt.Errorf("can't find authorizer named '%s'", name)
	}

	return a, nil
}

func (p *Project) GetSecondFactor(name string) (plugins.SecondFactor, error) {
	s, ok := p.secondFactors[name]
	if !ok || s == nil {
		return nil, fmt.Errorf("can't find second factor named '%s'", name)
	}

	return s, nil
}

func (p *Project) GetStorage(name string) (plugins.Storage, error) {
	s, ok := p.storages[name]
	if !ok || s == nil {
		return nil, fmt.Errorf("can't find storage named '%s'", name)
	}

	return s, nil
}

func (p *Project) GetKeyStorage(name string) (plugins.KeyStorage, error) {
	s, ok := p.keyStorages[name]
	if !ok || s == nil {
		return nil, fmt.Errorf("can't find key storage named '%s'", name)
	}

	return s, nil
}

func (p *Project) GetHasher(name string) (plugins.PWHasher, error) {
	h, ok := p.hashers[name]
	if !ok || h == nil {
		return nil, fmt.Errorf("can't find hasher named '%s'", name)
	}

	return h, nil
}

func (p *Project) GetSender(name string) (plugins.Sender, error) {
	s, ok := p.senders[name]
	if !ok || s == nil {
		return nil, fmt.Errorf("can't find sender named '%s'", name)
	}

	return s, nil
}

func (p *Project) GetCryptoKey(name string) (plugins.CryptoKey, error) {
	k, ok := p.cryptoKeys[name]
	if !ok || k == nil {
		return nil, fmt.Errorf("can't find crypto key named '%s'", name)
	}

	return k, nil
}

func (p *Project) GetServiceSignKey() (plugins.CryptoKey, error) {
	if p.service.signKey == nil {
		return nil, errors.New("cannot find service sign key")
	}
	return p.service.signKey, nil
}

func (p *Project) GetServiceEncKey() (plugins.CryptoKey, error) {
	if p.service.encKey == nil {
		return nil, errors.New("cannot find service enc key")
	}
	return p.service.encKey, nil
}

func (p *Project) GetServiceStorage() (plugins.Storage, error) {
	if p.service.storage == nil {
		return nil, errors.New("cannot find service storage")
	}
	return p.service.storage, nil
}

type App struct {
	name            string
	url             *url.URL
	pathPrefix      string
	authSessionExp  int
	identityManager identity.ManagerI
	authenticators  map[string]plugins.Authenticator
	authorizer      plugins.Authorizer
	secondFactor    plugins.SecondFactor
}

func (a *App) GetName() string {
	return a.name
}

func (a *App) GetUrl() (url.URL, error) {
	if a.url == nil {
		return url.URL{}, fmt.Errorf("can't find app url for app '%s'", a.name)
	}
	return *a.url, nil
}

func (a *App) GetPathPrefix() string {
	return a.pathPrefix
}

func (a *App) GetAuthSessionExp() int {
	return a.authSessionExp
}

func (a *App) GetIdentityManager() (identity.ManagerI, error) {
	if a.identityManager == nil {
		return nil, fmt.Errorf("can't find identity for app '%s'", a.name)
	}
	return a.identityManager, nil
}

func (a *App) GetAuthorizer() (plugins.Authorizer, error) {
	if a.authorizer == nil {
		return nil, fmt.Errorf("can't find authorizer for app '%s'", a.name)
	}
	return a.authorizer, nil
}

func (a *App) GetSecondFactor() (plugins.SecondFactor, error) {
	if a.secondFactor == nil {
		return nil, fmt.Errorf("can't find second factor for app '%s'", a.name)
	}
	return a.secondFactor, nil
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
