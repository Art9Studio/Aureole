package plugins

import (
	"aureole/internal/configs"
	"fmt"
)

var StorageRepo = createRepository()

type (
	StorageAdapter interface {
		Create(*configs.Storage) Storage
	}

	Storage interface {
		MetaDataGetter
		Set(k string, v interface{}, exp int) error
		Get(k string, v interface{}) (ok bool, err error)
		Delete(k string) error
		Exists(k string) (found bool, err error)
		Close() error
	}
)

func NewStorage(conf *configs.Storage) (Storage, error) {
	a, err := StorageRepo.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(StorageAdapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}
