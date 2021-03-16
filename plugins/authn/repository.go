package authn

import (
	"aureole/configs"
	ctxTypes "aureole/context/types"
	"aureole/plugins/authn/types"
	"fmt"
	"sync"
)

var (
	adapters   = make(map[string]Adapter)
	adaptersMU sync.Mutex
	projectCtx *ctxTypes.ProjectCtx
)

// Adapter defines methods for authentication plugins
type Adapter interface {
	// GetAuthnController returns desired authentication Controller depneds on the given config
	Get(config *configs.Authn, projectCtx *ctxTypes.ProjectCtx) (types.Controller, error)
}

func InitRepository(context *ctxTypes.ProjectCtx) {
	projectCtx = context
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
