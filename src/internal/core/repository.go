package core

import (
	"aureole/internal/configs"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

type (
	Type string

	Meta struct {
		Name string `yaml:"name"`
		Type Type   `yaml:"type"`
	}

	Claim[T Plugin] struct {
		Meta
		PluginCreate func(configs.PluginConfig) T
	}

	Repository[T Plugin] struct {
		pluginsMU sync.Mutex
		plugins   map[string]*Claim[T]
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
func (repo *Repository[T]) Get(name string, pluginType Type) (*Claim[T], error) {
	repo.pluginsMU.Lock()
	defer repo.pluginsMU.Unlock()
	path := buildPath(name, pluginType)
	if p, ok := repo.plugins[path]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("there is no pluginCreator for %s", name)
}

// Register registers Plugin
func (repo *Repository[T]) Register(metaYaml []byte, p func(configs.PluginConfig) T) Meta {
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

	repo.plugins[path] = &Claim[T]{
		Meta:         *meta,
		PluginCreate: p,
	}

	return *meta
}

func CreateRepository[T Plugin]() *Repository[T] {
	return &Repository[T]{
		plugins:   make(map[string]*Claim[T]),
		pluginsMU: sync.Mutex{},
	}
}

var AuthenticatorRepo = CreateRepository[Authenticator]()
var CryptoKeyRepo = CreateRepository[CryptoKey]()
var CryptoStorageRepo = CreateRepository[CryptoStorage]()
var IDManagerRepo = CreateRepository[IDManager]()
var IssuerRepo = CreateRepository[Issuer]()
var MFARepo = CreateRepository[MFA]()
var RootRepo= CreateRepository[RootPlugin]()
var SenderRepo = CreateRepository[Sender]()
var StorageRepo = CreateRepository[Storage]()

func CreatePlugin[T Plugin](repository *Repository[T], config configs.PluginConfig, pluginType Type) (T, error) {
	var empty T
	creator, err := repository.Get(config.Plugin, pluginType)
	if err != nil {
		return empty, err
	}

	plugin := creator.PluginCreate(config)

	return plugin, nil
}

//	if meta.Name == "" {
//		rndName, err := gonanoid.New(8)
//		if err != nil {
//			panic("could not generate random name")
//		}
//		meta.Name = rndName
//	}
