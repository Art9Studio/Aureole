package types

import (
	"aureole/internal/collections"
	authnTypes "aureole/internal/plugins/authn/types"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/core"
	ckeyTypes "aureole/internal/plugins/cryptokey/types"
	pwhasherTypes "aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	storageTypes "aureole/internal/plugins/storage/types"
)

type (
	ProjectCtx struct {
		APIVersion  string
		Apps        map[string]*App
		Collections map[string]*collections.Collection
		Storages    map[string]storageTypes.Storage
		Hashers     map[string]pwhasherTypes.PwHasher
		Senders     map[string]senderTypes.Sender
		CryptoKeys  map[string]ckeyTypes.CryptoKey
		Routes      []*core.Route
	}

	App struct {
		PathPrefix     string
		Authenticators []authnTypes.Authenticator
		Authorizers    map[string]authzTypes.Authorizer
	}
)
