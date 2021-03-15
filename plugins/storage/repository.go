package storage

import (
	"aureole/configs"
	ctxTypes "aureole/context/types"
	"fmt"
	"sync"
)

var (
	adapters   = make(map[string]Adapter)
	adaptersMU sync.Mutex
)

// Adapter define methods for storage plugins
type Adapter interface {
	Get(*configs.Storage) (ConnSession, error)

	//// OpenConfig attempts to establish a connection with a db by ConnConfig
	//OpenWithConfig(connConf ConnConfig) (ConnSession, error)
	//
	//// ParseUrl parses the connection url into ConnConfig struct
	//ParseUrl(connUrl string) (ConnConfig, error)
	//
	//// NewConfig creates new ConnConfig struct from the raw data, parsed from the config file
	//NewConfig(data map[string]interface{}) (ConnConfig, error)
}

func InitRepository(context *ctxTypes.ProjectCtx) {
	// For now, ignore project context.
}

// RegisterAdapter register storage adapter
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

// GetAdapter returns storage adapter if it exists
func GetAdapter(name string) (Adapter, error) {
	adaptersMU.Lock()
	defer adaptersMU.Unlock()

	if a, ok := adapters[name]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("can't find adapter named %s", name)
}
