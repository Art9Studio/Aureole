package core

import (
	"errors"
	"fmt"
	"regexp"
	"sync"

	"aureole/internal/configs"

	"gopkg.in/yaml.v3"
)

type (
	PluginType string
	PluginID   int

	Metadata struct {
		// lowercase name without spaces
		ShortName string `yaml:"name"`
		// human-readable name
		DisplayName string     `yaml:"display_name"`
		PluginID    PluginID   `yaml:"plugin_id"`
		Type        PluginType `yaml:"-"`
	}

	Claim[T Plugin] struct {
		Metadata
		PluginCreate func(configs.PluginConfig) T
	}

	Repository[T Plugin] struct {
		pluginsMU sync.Mutex
		plugins   map[string]*Claim[T]
		pluginIDs map[PluginID]struct{}
	}
)

const (
	TypeAuth          PluginType = "auth"
	TypeIssuer        PluginType = "issuer"
	TypeCryptoKey     PluginType = "crypto_key"
	TypeCryptoStorage PluginType = "crypto_storage"
	TypeIDManager     PluginType = "identity"
	TypeMFA           PluginType = "mfa"
	TypeRoot          PluginType = "root"
	TypeSender        PluginType = "sender"
	TypeStorage       PluginType = "storage"
	Invalid           PluginType = ""
)

// https://github.com/golang/go/issues/45380
func toPluginType[T Plugin]() (PluginType, error) {
	var p = new(T)
	switch any(p).(type) {
	case *Authenticator:
		return TypeAuth, nil
	case *Issuer:
		return TypeIssuer, nil
	case *CryptoKey:
		return TypeCryptoKey, nil
	case *CryptoStorage:
		return TypeCryptoStorage, nil
	case *IDManager:
		return TypeIDManager, nil
	case *MFA:
		return TypeMFA, nil
	case *RootPlugin:
		return TypeRoot, nil
	case *Sender:
		return TypeSender, nil
	case *Storage:
		return TypeStorage, nil
	default:
		return Invalid, errors.New("Invalid leave type")
	}
}

func buildPath(name string, pluginType PluginType) string {
	return fmt.Sprintf("%s.%s", pluginType, name)
}

func (repo *Repository[T]) Get(name string) (*Claim[T], error) {
	pluginType, err := toPluginType[T]()
	if err != nil {
		return nil, err
	}
	repo.pluginsMU.Lock()
	defer repo.pluginsMU.Unlock()
	path := buildPath(name, pluginType)
	if p, ok := repo.plugins[path]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("there is no pluginCreator for %s", name)
}

// Register registers Plugin
func (repo *Repository[T]) Register(metaYaml []byte, p func(configs.PluginConfig) T) Metadata {
	var meta = &Metadata{}
	err := yaml.Unmarshal(metaYaml, meta)
	if err != nil {
		return *meta
	}

	meta.Type, err = toPluginType[T]()
	if err != nil {
		panic("invalid plugin type")
	}
	// todo: validate name. it should has no spaces, should be lowercase
	if !regexShort.MatchString(meta.ShortName) {
		panic("name for a Plugin with type " + meta.Type + " wasn't passed")
	}
	// todo: validate display name. it should start with upper case
	if !regexDisplay.MatchString(meta.DisplayName) {
		panic("display name for a Plugin " + meta.ShortName + " wasn't passed")
	}
	repo.pluginsMU.Lock()
	defer repo.pluginsMU.Unlock()

	path := buildPath(meta.ShortName, meta.Type)
	if _, ok := repo.plugins[path]; ok {
		panic("multiple Register call for Plugin " + path)
	}

	id := meta.PluginID
	if _, ok := repo.pluginIDs[id]; ok {
		panic(fmt.Sprintf("multiple Register call for PluginID %d", id))
	}

	repo.plugins[path] = &Claim[T]{
		Metadata:     *meta,
		PluginCreate: p,
	}
	repo.pluginIDs[id] = struct{}{}

	return *meta
}

func CreateRepository[T Plugin]() *Repository[T] {
	return &Repository[T]{
		plugins:   make(map[string]*Claim[T]),
		pluginIDs: make(map[PluginID]struct{}),
		pluginsMU: sync.Mutex{},
	}
}

var AuthenticatorRepo = CreateRepository[Authenticator]()
var CryptoKeyRepo = CreateRepository[CryptoKey]()
var CryptoStorageRepo = CreateRepository[CryptoStorage]()
var IDManagerRepo = CreateRepository[IDManager]()
var IssuerRepo = CreateRepository[Issuer]()
var MFARepo = CreateRepository[MFA]()
var RootRepo = CreateRepository[RootPlugin]()
var SenderRepo = CreateRepository[Sender]()
var StorageRepo = CreateRepository[Storage]()

var regexShort = regexp.MustCompile("^[a-z]+$")
var regexDisplay = regexp.MustCompile("^[A-Z][a-z]+$")

func CreatePlugin[T Plugin](repository *Repository[T], config configs.PluginConfig) (T, error) {
	var empty T

	creator, err := repository.Get(config.Plugin)
	if err != nil {
		return empty, err
	}

	plugin := creator.PluginCreate(config)

	return plugin, nil
}

//	if meta.ShortName == "" {
//		rndName, err := gonanoid.New(8)
//		if err != nil {
//			panic("could not generate random name")
//		}
//		meta.ShortName = rndName
//	}
