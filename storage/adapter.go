package storage

import (
	"fmt"
	"sync"
)

var (
	adapters   = make(map[string]Adapter)
	adaptersMU sync.Mutex
)

// Adapter define methods for storage adapters
type Adapter interface {
	// OpenConfig attempts to establish a connection with a db by ConnectionConfig
	OpenConfig(connConf ConnectionConfig) (Session, error)

	ParseUrl(connUrl string) (ConnectionConfig, error)

	NewConfig(data map[string]interface{}) (ConnectionConfig, error)
}

// RegisterAdapter register storage adapter
func RegisterAdapter(name string, a Adapter) {
	adaptersMU.Lock()
	defer adaptersMU.Unlock()

	if name == "" {
		panic("adapter name can't be empty")
	}

	if _, ok := adapters[name]; ok {
		panic("multiply RegisterAdapter call for adapter " + name)
	}

	adapters[name] = a
}

// GetAdapter returns storage adapter if it exists
func GetAdapter(name string) (Adapter, error) {
	adaptersMU.Lock()
	defer adaptersMU.Unlock()

	if a, ok := adapters[name]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("can't find adapter named %s", name)
}
