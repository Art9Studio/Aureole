package _interface

import (
	"aureole/internal/identity"
	"aureole/internal/plugins/2fa/types"
	authzT "aureole/internal/plugins/authz/types"
	cryptoKeyT "aureole/internal/plugins/cryptokey/types"
	kstorageT "aureole/internal/plugins/kstorage/types"
	pwhasherT "aureole/internal/plugins/pwhasher/types"
	senderT "aureole/internal/plugins/sender/types"
	storageT "aureole/internal/plugins/storage/types"
	"net/url"
)

type (
	ProjectState interface {
		GetAPIVersion() string
		GetPingPath() string
		IsTestRun() bool
		GetApp(name string) (AppState, error)
		GetAuthorizer(name string) (authzT.Authorizer, error)
		GetSecondFactor(name string) (types.SecondFactor, error)
		GetStorage(name string) (storageT.Storage, error)
		GetKeyStorage(name string) (kstorageT.KeyStorage, error)
		GetHasher(name string) (pwhasherT.PwHasher, error)
		GetSender(name string) (senderT.Sender, error)
		GetCryptoKey(name string) (cryptoKeyT.CryptoKey, error)
		GetServiceSignKey() (cryptoKeyT.CryptoKey, error)
		GetServiceEncKey() (cryptoKeyT.CryptoKey, error)
		GetServiceStorage() (storageT.Storage, error)
	}

	AppState interface {
		GetName() string
		GetUrl() (url.URL, error)
		GetPathPrefix() string
		GetIdentityManager() (identity.ManagerI, error)
		GetAuthorizer() (authzT.Authorizer, error)
		GetSecondFactor() (types.SecondFactor, error)
		Filter(data, filter map[string]string) (bool, error)
	}
)
