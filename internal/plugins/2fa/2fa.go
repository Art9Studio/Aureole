package second_factor

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/2fa/types"
	"fmt"
)

func New(conf *configs.SecondFactor) (types.SecondFactor, error) {
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
