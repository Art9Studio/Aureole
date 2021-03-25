package types

import (
	"aureole/internal/collections"
	authnTypes "aureole/internal/plugins/authn/types"
	authzTypes "aureole/internal/plugins/authz/types"
	pwhasherTypes "aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	storageTypes "aureole/internal/plugins/storage/types"
)

type (
	ProjectCtx struct {
		APIVersion  string
		Apps        map[string]App
		Collections map[string]*collections.Collection
		Storages    map[string]storageTypes.Storage
		Hashers     map[string]pwhasherTypes.PwHasher
		Senders     map[string]senderTypes.Sender
	}

	App struct {
		PathPrefix       string
		AuthnControllers []authnTypes.Controller
		Authorizers      []authzTypes.Authorizer
	}
)
