package core

import (
	"aureole/internal/plugins"
	"net/url"
)

type project struct {
	apiVersion string
	testRun    bool
	pingPath   string
	apps       map[string]*app
}

type (
	app struct {
		name           string
		url            *url.URL
		pathPrefix     string
		authSessionExp int
		service        service
		authenticators map[string]plugins.Authenticator
		authorizer     plugins.Authorizer
		secondFactors  map[string]plugins.SecondFactor
		idManager      plugins.IDManager
		storages       map[string]plugins.Storage
		keyStorages    map[string]plugins.KeyStorage
		hashers        map[string]plugins.PWHasher
		senders        map[string]plugins.Sender
		cryptoKeys     map[string]plugins.CryptoKey
		admins         map[string]plugins.Admin
	}

	service struct {
		signKey plugins.CryptoKey
		encKey  plugins.CryptoKey
		storage plugins.Storage
	}
)

func (a *app) getServiceSignKey() (plugins.CryptoKey, bool) {
	if a.service.signKey == nil {
		return nil, false
	}
	return a.service.signKey, true
}

func (a *app) getServiceEncKey() (plugins.CryptoKey, bool) {
	if a.service.encKey == nil {
		return nil, false
	}
	return a.service.encKey, true
}

func (a *app) getServiceStorage() (plugins.Storage, bool) {
	if a.service.storage == nil {
		return nil, false
	}
	return a.service.storage, true
}

func (a *app) getIDManager() (plugins.IDManager, bool) {
	if a.idManager == nil {
		return nil, false
	}
	return a.idManager, true
}

func (a *app) getAuthorizer() (plugins.Authorizer, bool) {
	if a.authorizer == nil {
		return nil, false
	}
	return a.authorizer, true
}

func (a *app) getSecondFactors() (map[string]plugins.SecondFactor, bool) {
	if a.secondFactors == nil {
		return nil, false
	}
	return a.secondFactors, true
}

func (a *app) getStorage(name string) (plugins.Storage, bool) {
	s, ok := a.storages[name]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}

func (a *app) getKeyStorage(name string) (plugins.KeyStorage, bool) {
	s, ok := a.keyStorages[name]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}

func (a *app) getHasher(name string) (plugins.PWHasher, bool) {
	h, ok := a.hashers[name]
	if !ok || h == nil {
		return nil, false
	}
	return h, true
}

func (a *app) getSender(name string) (plugins.Sender, bool) {
	s, ok := a.senders[name]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}

func (a *app) getCryptoKey(name string) (plugins.CryptoKey, bool) {
	k, ok := a.cryptoKeys[name]
	if !ok || k == nil {
		return nil, false
	}
	return k, true
}
