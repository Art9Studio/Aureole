package types

import (
	"gouth/adapters/authn/types"
	"gouth/adapters/pwhasher"
	"gouth/adapters/storage"
	"gouth/collections"
	"gouth/configs"
)

type ProjectCtx struct {
	APIVersion  string
	Apps        map[string]App
	Collections map[string]collections.Collection
	Storages    map[string]storage.ConnSession
	Hashers     map[string]pwhasher.PwHasher
}

type App struct {
	PathPrefix       string
	AuthnControllers []types.Controller
}

type AuthnConfig struct {
	PathPrefix string
	Type       types.Type
	Config     configs.RawConfig
}

type AuthZConfig struct {
	Type   string
	Config configs.RawConfig
}
