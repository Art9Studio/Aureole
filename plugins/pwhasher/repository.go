package pwhasher

import (
	"aureole/configs"
	ctxTypes "aureole/context/types"
	"aureole/plugins/pwhasher/types"
	"fmt"
	"sync"
)

var (
	adapters   = make(map[string]Adapter)
	adaptersMU sync.Mutex
)

// Adapter defines methods for pwhasher plugins
type Adapter interface {
	//GetPwHasher returns desired PwHasher depends on the given config
	Get(*configs.Hasher) (types.PwHasher, error)
}

func InitRepository(context *ctxTypes.ProjectCtx) {
	// For now, ignore project context. At 95% we dont need pass the Project context to hasher
}

// RegisterAdapter register pwhasher adapter
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

// GetAdapter returns pwhasher adapter if it exists
func GetAdapter(name string) (Adapter, error) {
	adaptersMU.Lock()
	defer adaptersMU.Unlock()

	if a, ok := adapters[name]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("can't find adapter named %s", name)
}
