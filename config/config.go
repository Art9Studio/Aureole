package config

import (
	"gouth/authN"
	"gouth/authN/types"
	"gouth/pwhasher"
	"gouth/storage"
)

type Project struct {
	APIVersion  string
	Apps        []App
	Collections map[string]interface{}
	Storages    map[string]storage.ConnSession
	Hashers     map[string]pwhasher.PwHasher
}

type App struct {
	PathPrefix       string
	AuthNControllers []authN.Controller
	AuthNConfigs     []AuthNConfig
}

type AuthNConfig struct {
	PathPrefix string
	Type       types.Type
	Config     RawConfig
}

type AuthZConfig struct {
	Type   string
	Config RawConfig
}
