package plugins

import (
	"aureole/internal/configs"
	"fmt"
)

var AdminRepo = createRepository()

type (
	// AdminAdapter defines methods for admin adapter
	AdminAdapter interface {
		// Create returns desired admin plugin depends on the given config
		Create(admin *configs.Admin) Admin
	}

	Admin interface {
		MetaDataGetter
	}
)

func NewAdmin(conf *configs.Admin) (Admin, error) {
	a, err := AdminRepo.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(AdminAdapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}
