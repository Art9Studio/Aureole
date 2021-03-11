package config

import (
	"gouth/authN/types"
	"gouth/pwhasher"
	"gouth/storage"
)

type Project struct {
	APIVersion  string
	Apps        []App
	Collections []interface{}
	Storages    map[string]storage.ConnSession
	Hashers     pwhasher.PwHasher
}

type App struct {
	PathPrefix string
	AuthN      []AuthNConfig
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
