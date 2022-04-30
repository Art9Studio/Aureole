package plugins

import (
	"aureole/internal/configs"
)

type (
	CryptoStorageCreator interface {
		Create(*configs.PluginConfig) CryptoStorage
	}

	CryptoStorage interface {
		MetaDataGetter
		Read(v *[]byte) (ok bool, err error)
		Write(v []byte) error
	}
)
