package types

import (
	"aureole/collections"
	"aureole/configs"
	"aureole/plugins/authn/types"
	"aureole/plugins/pwhasher"
	"aureole/plugins/storage"
)

type (
	ProjectCtx struct {
		APIVersion  string
		Apps        map[string]App
		Collections map[string]*collections.Collection
		Storages    map[string]storage.ConnSession
		Hashers     map[string]pwhasher.PwHasher
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
