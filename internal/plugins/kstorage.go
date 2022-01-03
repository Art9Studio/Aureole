package plugins

import (
	"aureole/internal/configs"
	"fmt"
)

var KeyStorageRepo = createRepository()

type (
	KeyStorageAdapter interface {
		Create(*configs.KeyStorage) KeyStorage
	}

	KeyStorage interface {
		MetaDataGetter
		Read(v *[]byte) (ok bool, err error)
		Write(v []byte) error
	}
)

func NewKeyStorage(conf *configs.KeyStorage) (KeyStorage, error) {
	a, err := KeyStorageRepo.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(KeyStorageAdapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}
