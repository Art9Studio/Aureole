package _interface

import (
	"aureole/internal/collections"
	"aureole/internal/identity"
	authzTypes "aureole/internal/plugins/authz/types"
	cryptoKeyTypes "aureole/internal/plugins/cryptokey/types"
	pwhasherTypes "aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"net/url"
)

type (
	ProjectState interface {
		IsTestRun() bool
		GetCollection(name string) (*collections.Collection, error)
		GetStorage(name string) (storageTypes.Storage, error)
		GetHasher(name string) (pwhasherTypes.PwHasher, error)
		GetSender(name string) (senderTypes.Sender, error)
		GetCryptoKey(name string) (cryptoKeyTypes.CryptoKey, error)
	}

	AppState interface {
		GetName() string
		GetUrl() (*url.URL, error)
		GetPathPrefix() string
		GetIdentity() (*identity.Identity, error)
		GetAuthorizer(name string) (authzTypes.Authorizer, error)
	}
)