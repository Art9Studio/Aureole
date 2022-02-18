package plugins

import (
	"aureole/internal/configs"
	"fmt"
)

var CryptoStorageRepo = createRepository()

type (
	CryptoStorageAdapter interface {
		Create(*configs.CryptoStorage) CryptoStorage
	}

	CryptoStorage interface {
		MetaDataGetter
		Read(v *[]byte) (ok bool, err error)
		Write(v []byte) error
	}
)

func NewCryptoStorage(conf *configs.CryptoStorage) (CryptoStorage, error) {
	a, err := CryptoStorageRepo.Get(conf.Type)
	if err != nil {
		return nil, err
	}

	adapter, ok := a.(CryptoStorageAdapter)
	if !ok {
		return nil, fmt.Errorf("trying to cast adapter was failed: %v", err)
	}

	return adapter.Create(conf), nil
}
