package plugins

import (
	"aureole/internal/configs"
	"fmt"
)

var UIRepo = createRepository()

type (
	// UIAdapter defines methods for ui adapter
	UIAdapter interface {
		// Create returns desired ui plugin depends on the given config
		Create(ui *configs.UI) UI
	}

	UI interface {
		MetaDataGetter
	}
)

func NewUI(conf *configs.UI) (UI, error) {
	a, err := UIRepo.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(UIAdapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}
