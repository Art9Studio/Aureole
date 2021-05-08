package collections

import (
	"fmt"
	"sync"
)

type repository struct {
	collections map[string]*CollectionType
	mu          sync.Mutex
}

var Repository = &repository{
	collections: make(map[string]*CollectionType),
	mu:          sync.Mutex{},
}

// Get returns CollectionType if it exists
func (repo *repository) Get(name string) (*CollectionType, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if a, ok := repo.collections[name]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("can't find CollectionType named %s", name)
}

// Register registers CollectionType
func (repo *repository) Register(c *CollectionType) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if c.Name == "" {
		panic("collectionType name can't be empty")
	}

	if _, ok := repo.collections[c.Name]; ok {
		panic("multiply Register call for adapter " + c.Name)
	}

	repo.collections[c.Name] = c
}
