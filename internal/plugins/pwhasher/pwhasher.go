package pwhasher

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/pwhasher/types"
	"fmt"
)

// New returns desired pwHasher depends on the given config
func New(conf *configs.PwHasher) (types.PwHasher, error) {
	a, err := Repository.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(Adapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}
