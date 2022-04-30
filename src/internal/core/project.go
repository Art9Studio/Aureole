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
		internal       internal
		authenticators map[string]plugins.Authenticator
		issuer         plugins.Issuer
		mfa            map[string]plugins.MFA
		idManager      plugins.IDManager
		storages       map[string]plugins.Storage
		cryptoStorages map[string]plugins.CryptoStorage
		senders        map[string]plugins.Sender
		cryptoKeys     map[string]plugins.CryptoKey
		rootPlugins    map[string]plugins.RootPlugin
	}

	internal struct {
		signKey plugins.CryptoKey
		encKey  plugins.CryptoKey
		storage plugins.Storage
	}
)

func (a *app) getServiceSignKey() (plugins.CryptoKey, bool) {
	if a.internal.signKey == nil {
		return nil, false
	}
	return a.internal.signKey, true
}

func (a *app) getServiceEncKey() (plugins.CryptoKey, bool) {
	if a.internal.encKey == nil {
		return nil, false
	}
	return a.internal.encKey, true
}

func (a *app) getServiceStorage() (plugins.Storage, bool) {
	if a.internal.storage == nil {
		return nil, false
	}
	return a.internal.storage, true
}

func (a *app) getIDManager() (plugins.IDManager, bool) {
	if a.idManager == nil {
		return nil, false
	}
	return a.idManager, true
}

func (a *app) getIssuer() (plugins.Issuer, bool) {
	if a.issuer == nil {
		return nil, false
	}
	return a.issuer, true
}

func (a *app) getSecondFactors() (map[string]plugins.MFA, bool) {
	if a.mfa == nil || len(a.mfa) == 0 {
		return nil, false
	}
	return a.mfa, true
}

func (a *app) getStorage(name string) (plugins.Storage, bool) {
	s, ok := a.storages[name]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}

func (a *app) getCryptoStorage(name string) (plugins.CryptoStorage, bool) {
	s, ok := a.cryptoStorages[name]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
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
