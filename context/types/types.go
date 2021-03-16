package types

import (
	"aureole/collections"
	"aureole/configs"
	authnTypes "aureole/plugins/authn/types"
	pwhasherTypes "aureole/plugins/pwhasher/types"
	storageTypes "aureole/plugins/storage/types"
)

type (
	ProjectCtx struct {
		APIVersion  string
		Apps        map[string]App
		Collections map[string]*collections.Collection
		Storages    map[string]storageTypes.Storage
		Hashers     map[string]pwhasherTypes.PwHasher
	}

	App struct {
		PathPrefix       string
		AuthnControllers []authnTypes.Controller
	}

	AuthZConfig struct {
		Type   string
		Config configs.RawConfig
	}
)
