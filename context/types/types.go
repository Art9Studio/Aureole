package types

import (
	"gouth/adapters/authn/types"
	"gouth/adapters/pwhasher"
	"gouth/adapters/storage"
	"gouth/collections"
	"gouth/configs"
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
