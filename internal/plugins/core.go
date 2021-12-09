package plugins

import (
	"fmt"
	"sync"
)

type (
	MetaDataGetter interface {
		GetMetaData() Meta
	}

	Meta struct {
		Type string
		Name string
		ID   string
	}

	adapter interface {
	}

	repository struct {
		adaptersMU sync.Mutex
		adapters   map[string]adapter
	}
)

// Get returns kstorage adapter if it exists
func (repo *repository) Get(name string) (adapter, error) {
	repo.adaptersMU.Lock()
	defer repo.adaptersMU.Unlock()

	if a, ok := repo.adapters[name]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("can't find adapter named %s", name)
}

// Register registers adapter
func (repo *repository) Register(name string, a adapter) {
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

func createRepository() *repository {
	return &repository{
		adapters:   make(map[string]adapter),
		adaptersMU: sync.Mutex{},
	}
}
