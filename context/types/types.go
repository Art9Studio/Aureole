package types

import (
	"gouth/adapters/authn/types"
	"gouth/adapters/pwhasher"
	"gouth/adapters/storage"
	"gouth/config"
)

type ProjectCtx struct {
	APIVersion  string
	Apps        []App
	Collections map[string]interface{}
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
	Config     config.RawConfig
}

type AuthZConfig struct {
	Type   string
	Config config.RawConfig
}
