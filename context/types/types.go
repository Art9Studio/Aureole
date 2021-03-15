package types

import (
	"aureole/collections"
	"aureole/configs"
	"aureole/plugins/authn/types"
	types2 "aureole/plugins/pwhasher/types"
	"aureole/plugins/storage"
)

type (
	ProjectCtx struct {
		APIVersion  string
		Apps        map[string]App
		Collections map[string]*collections.Collection
		Storages    map[string]storage.ConnSession
		Hashers     map[string]types2.PwHasher
	}

	App struct {
		PathPrefix       string
		AuthnControllers []types.Controller
	}

	AuthZConfig struct {
		Type   string
		Config configs.RawConfig
	}
)
