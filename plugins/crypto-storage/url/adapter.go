package url

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const adapterName = "url"

func init() {
	plugins.CryptoStorageRepo.Register(adapterName, adapter{})
}

type adapter struct {
}

func (adapter) Create(conf *configs.CryptoStorage) plugins.CryptoStorage {
	return &cryptoStorage{rawConf: conf}
}
