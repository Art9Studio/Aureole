package _interface

import (
	"aureole/internal/identity"
	authzT "aureole/internal/plugins/authz/types"
	cryptoKeyT "aureole/internal/plugins/cryptokey/types"
	kstorageT "aureole/internal/plugins/kstorage/types"
	pwhasherT "aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	storageT "aureole/internal/plugins/storage/types"
	"net/url"
)

type (
	ProjectState interface {
		PluginsState
		GetServiceKey() (cryptoKeyT.CryptoKey, error)
		GetServiceStorage() (storageT.Storage, error)
	}

	PluginsState interface {
		IsTestRun() bool
		GetKeyStorage(name string) (kstorageT.KeyStorage, error)
		GetStorage(name string) (storageT.Storage, error)
		GetHasher(name string) (pwhasherT.PwHasher, error)
		GetSender(name string) (senderTypes.Sender, error)
		GetCryptoKey(name string) (cryptoKeyT.CryptoKey, error)
	}

	AppState interface {
		GetName() string
		GetUrl() (url.URL, error)
		GetPathPrefix() string
		GetIdentityManager() (identity.ManagerI, error)
		GetAuthorizer() (authzT.Authorizer, error)
		Filter(data, filter map[string]string) (bool, error)
	}
)
