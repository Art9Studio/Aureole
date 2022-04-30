package plugins

import (
	"aureole/internal/configs"
	"fmt"
	"github.com/go-openapi/spec"
	"gopkg.in/yaml.v3"
	"sync"
)

type (
	PluginCreator interface {
		Create(configs.PluginConfig) Plugin
	}

	Plugin interface {
		MetaDataGetter
	}

	MetaDataGetter interface {
		GetMetaData() Meta
	}

	PluginType string

	Meta struct {
		Name string     `yaml:"name"`
		Type PluginType `yaml:"type"`
	}

	OpenAPISpecGetter interface {
		GetHandlersSpec() (*spec.Paths, spec.Definitions)
	}

	PluginClaim struct {
		Meta
		PluginCreator
	}

	repository struct {
		pluginsMU sync.Mutex
		plugins   map[string]PluginClaim
	}
)

const (
	PluginTypeAuth          PluginType = "auth"
	PluginTypeIssuer        PluginType = "issuer"
	PluginTypeCryptoKey     PluginType = "crypto_key"
	PluginTypeCryptoStorage PluginType = "crypto_storage"
	PluginTypeIDManager     PluginType = "identity"
	PluginTypeMFA           PluginType = "mfa"
	PluginTypeRoot          PluginType = "root"
	PluginTypeSender        PluginType = "sender"
	PluginTypeStorage       PluginType = "storage"
)

var Repo = createRepository()

func buildPath(name string, pluginType PluginType) string {
	return fmt.Sprintf("%s.%s", pluginType, name)
}

// Get returns kstorage plugin if it exists
func (repo *repository) Get(name string, pluginType PluginType) (PluginCreator, error) {
	repo.pluginsMU.Lock()
	defer repo.pluginsMU.Unlock()
	path := buildPath(name, pluginType)
	if p, ok := repo.plugins[path]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("there is no pluginCreator for %s", name)
}

// Register registers plugin
func (repo *repository) Register(metaYaml []byte, p PluginCreator) Meta {
	var meta = &Meta{}
	err := yaml.Unmarshal(metaYaml, meta)
	if err != nil {
		return *meta
	}
	if meta.Name == "" {
		panic("name for a plugin wasn't passed")
	}
	repo.pluginsMU.Lock()
	defer repo.pluginsMU.Unlock()

	// todo: validate meta
	path := buildPath(meta.Name, meta.Type)
	if _, ok := repo.plugins[path]; ok {
		panic("multiple Register call for plugin " + path)
	}

	repo.plugins[path] = PluginClaim{
		Meta:          *meta,
		PluginCreator: p,
	}

	return *meta
}

func createRepository() *repository {
	return &repository{
		plugins:   make(map[string]PluginClaim),
		pluginsMU: sync.Mutex{},
	}
}

func CreatePlugin[T Plugin](config configs.PluginConfig, pluginType PluginType) (T, error) {
	var empty T
	creator, err := Repo.Get(config.Plugin, pluginType)
	if err != nil {
		return empty, err
	}

	abstractPlugin := creator.Create(config)

	plugin, ok := abstractPlugin.(T)

	if !ok {
		return empty, fmt.Errorf("trying to cast plugin was failed")
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
