package pwhasher

import (
	"aureole/configs"
	"aureole/internal/plugins/pwhasher/types"
	"fmt"
)

// New returns desired PwHasher depends on the given config
func New(conf *configs.PwHasher) (types.PwHasher, error) {
	a, err := Repository.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := interface{}(a).(Adapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf)
}
