package sender

import (
	"aureole/configs"
	"aureole/internal/plugins/sender/types"
	"fmt"
)

// New returns desired messenger depends on the given config
func New(conf *configs.Sender) (types.Sender, error) {
	a, err := Repository.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(Adapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf)
}
