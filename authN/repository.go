package authN

import (
	"fmt"
	"gouth/config"
	"sync"
)

var (
	adapters       map[string]Adapter
	adaptersMU     sync.Mutex
	projectContext *config.Project
)

// Adapter defines methods for authentication adapters
type Adapter interface {
	// GetAuthNController returns desired authentication controller depends on the given config
	GetAuthNController(pathPrefix string, config *config.RawConfig, projectContext *config.Project) (Controller, error)
}

// RegisterAdapter register authentication adapter
func RegisterAdapter(name string, a Adapter) {
	adaptersMU.Lock()
	defer adaptersMU.Unlock()

	if name == "" {
		panic("adapter Name can't be empty")
	}

	if _, ok := adapters[name]; ok {
		panic("multiply RegisterAdapter call for adapter " + name)
	}

	adapters[name] = a
}

// GetAdapter returns authentication adapter if it exists
func GetAdapter(name string) (Adapter, error) {
	adaptersMU.Lock()
	defer adaptersMU.Unlock()

	if a, ok := adapters[name]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("can't find adapter named %s", name)
}

func InitRepository(context *config.Project) {
	adapters = make(map[string]Adapter)
	projectContext = context
}
