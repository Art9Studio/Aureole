package authn

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/authn/types"
	"fmt"
)

// New returns desired Authenticator depends on the given config
func New(conf *configs.Authn) (types.Authenticator, error) {
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
