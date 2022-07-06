package core

import (
	"aureole/internal/configs"
	"fmt"
	"gopkg.in/yaml.v3"
	"sync"
)

type (
	Type string

	Meta struct {
		Name string `yaml:"name"`
		Type Type   `yaml:"type"`
	}

	PluginCreate = func(configs.PluginConfig) Plugin

	AuthPluginCreate = func(configs.PluginConfig) Authenticator

	Claim struct {
		Meta
		PluginCreate
	}

	Repository struct {
		pluginsMU sync.Mutex
		plugins   map[string]*Claim
	}
)

const (
	TypeAuth          Type = "auth"
	TypeIssuer        Type = "issuer"
	TypeCryptoKey     Type = "crypto_key"
	TypeCryptoStorage Type = "crypto_storage"
	TypeIDManager     Type = "identity"
	TypeMFA           Type = "mfa"
	TypeRoot          Type = "root"
	TypeSender        Type = "sender"
	TypeStorage       Type = "storage"
)

func buildPath(name string, pluginType Type) string {
	return fmt.Sprintf("%s.%s", pluginType, name)
}

// Get returns kstorage Plugin if it exists
func (repo *Repository) Get(name string, pluginType Type) (*Claim, error) {
	repo.pluginsMU.Lock()
	defer repo.pluginsMU.Unlock()
	path := buildPath(name, pluginType)
	if p, ok := repo.plugins[path]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("there is no pluginCreator for %s", name)
}

// Register registers Plugin
func (repo *Repository) Register(metaYaml []byte, p PluginCreate) Meta {
	var meta = &Meta{}
	err := yaml.Unmarshal(metaYaml, meta)
	if err != nil {
		return *meta
	}
	if meta.Name == "" {
		panic("name for a Plugin wasn't passed")
	}
	repo.pluginsMU.Lock()
	defer repo.pluginsMU.Unlock()

	// todo: validate meta
	path := buildPath(meta.Name, meta.Type)
	if _, ok := repo.plugins[path]; ok {
		panic("multiple Register call for Plugin " + path)
	}

	repo.plugins[path] = &Claim{
		Meta:         *meta,
		PluginCreate: p,
	}

	return *meta
}

func CreateRepository() *Repository {
	return &Repository{
		plugins:   make(map[string]*Claim),
		pluginsMU: sync.Mutex{},
	}
}

var Repo = CreateRepository()

func CreatePlugin[T Plugin](repository *Repository, config configs.PluginConfig, pluginType Type) (T, error) {
	var empty T
	creator, err := repository.Get(config.Plugin, pluginType)
	if err != nil {
		return empty, err
	}

	abstractPlugin := creator.PluginCreate(config)

	plugin, ok := abstractPlugin.(T)

	if !ok {
		return empty, fmt.Errorf("trying to cast Plugin was failed")
	}

	return plugin, nil
}

//	if meta.Name == "" {
//		rndName, err := gonanoid.New(8)
//		if err != nil {
//			panic("could not generate random name")
//		}
//		meta.Name = rndName
//	}
