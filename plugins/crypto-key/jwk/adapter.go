package jwk

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "jwk"

// init initializes package by register adapter
func init() {
	plugins.CryptoKeyRepo.Register(adapterName, adapter{})
}

// adapter represents adapter for jwk
type adapter struct {
}

func (adapter) Create(conf *configs.CryptoKey) plugins.CryptoKey {
	return &cryptoKey{rawConf: conf}
}
