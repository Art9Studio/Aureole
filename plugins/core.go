package plugins

import (
	ctxTypes "aureole/context/types"
	"fmt"
	"sync"
)

type Adapter interface {
}

type Repository struct {
	adapters   map[string]Adapter
	adaptersMU sync.Mutex
	ProjectCtx *ctxTypes.ProjectCtx
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

// Register register storage adapter
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

func InitRepository() *Repository {
	return &Repository{
		adapters:   make(map[string]Adapter),
		adaptersMU: sync.Mutex{},
	}
}
