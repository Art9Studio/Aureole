package core

import (
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
		authenticators map[string]Authenticator
		issuer         Issuer
		mfa            map[string]MFA
		idManager      IDManager
		storages       map[string]Storage
		cryptoStorages map[string]CryptoStorage
		senders        map[string]Sender
		cryptoKeys     map[string]CryptoKey
		rootPlugins    map[string]RootPlugin
		scratchCode    struct {
			Num      int    `mapstructure:"num" json:"num"`
			Alphabet string `mapstructure:"alphabet" json:"alphabet"`
		} `mapstructure:"scratch_code" json:"scratch_code"`
	}

	internal struct {
		signKey CryptoKey
		encKey  CryptoKey
		storage Storage
	}
)

func (a *app) getServiceSignKey() (CryptoKey, bool) {
	if a.internal.signKey == nil {
		return nil, false
	}
	return a.internal.signKey, true
}

func (a *app) getServiceEncKey() (CryptoKey, bool) {
	if a.internal.encKey == nil {
		return nil, false
	}
	return a.internal.encKey, true
}

func (a *app) getServiceStorage() (Storage, bool) {
	if a.internal.storage == nil {
		return nil, false
	}
	return a.internal.storage, true
}

func (a *app) getIDManager() (IDManager, bool) {
	if a.idManager == nil {
		return nil, false
	}
	return a.idManager, true
}

func (a *app) getIssuer() (Issuer, bool) {
	if a.issuer == nil {
		return nil, false
	}
	return a.issuer, true
}

func (a *app) getSecondFactors() (map[string]MFA, bool) {
	if a.mfa == nil || len(a.mfa) == 0 {
		return nil, false
	}
	return a.mfa, true
}

func (a *app) getStorage(name string) (Storage, bool) {
	s, ok := a.storages[name]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}

func (a *app) getCryptoStorage(name string) (CryptoStorage, bool) {
	s, ok := a.cryptoStorages[name]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}

func (a *app) getSender(name string) (Sender, bool) {
	s, ok := a.senders[name]
	if !ok || s == nil {
		return nil, false
	}
	return s, true
}

func (a *app) getCryptoKey(name string) (CryptoKey, bool) {
	k, ok := a.cryptoKeys[name]
	if !ok || k == nil {
		return nil, false
	}
	return k, true
}
