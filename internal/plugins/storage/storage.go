package storage

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/storage/types"
	"fmt"
)

// New returns desired storage depends on the given config
func New(conf *configs.Storage) (types.Storage, error) {
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
