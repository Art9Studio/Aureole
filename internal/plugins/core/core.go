package core

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type Adapter interface {
}

type Route struct {
	Method  string
	Path    string
	Handler func(*fiber.Ctx) error
}

type Repository struct {
	adapters   map[string]Adapter
	adaptersMU sync.Mutex
	PluginApi  *PluginApi
}

// Get returns storage adapter if it exists
func (repo *Repository) Get(name string) (Adapter, error) {
	repo.adaptersMU.Lock()
	defer repo.adaptersMU.Unlock()

	if a, ok := repo.adapters[name]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("can't find adapter named %s", name)
}

// Register registers adapter
func (repo *Repository) Register(name string, a Adapter) {
	repo.adaptersMU.Lock()
	defer repo.adaptersMU.Unlock()

	if name == "" {
		panic("adapter Name can't be empty")
	}

	if _, ok := repo.adapters[name]; ok {
		panic("multiply Register call for adapter " + name)
	}

	repo.adapters[name] = a
}

func CreateRepository() *Repository {
	return &Repository{
		adapters:   make(map[string]Adapter),
		adaptersMU: sync.Mutex{},
		PluginApi:  &pluginApi,
	}
}
