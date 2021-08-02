package admin

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/admin/types"
	"fmt"
)

func New(conf *configs.Admin) (types.Admin, error) {
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
