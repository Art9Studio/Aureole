package authz

import (
	"aureole/configs"
	"aureole/internal/plugins/authz/types"
	"fmt"
)

// New returns desired authorizer depends on the given config
func New(conf *configs.Authz) (types.Authorizer, error) {
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
