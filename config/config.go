package config

import (
	"gouth/authN"
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
	Path   string
	Type   authN.Type
	Config RawConfig
}

type AuthZConfig struct {
	Type   string
	Config RawConfig
}
