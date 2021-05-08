package cryptokey

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/core"
	"aureole/internal/plugins/cryptokey/types"
)

var Repository = core.CreateRepository()

// Adapter defines methods for authentication plugins
type Adapter interface {
	// Create returns desired crypto key depends on the given config
	Create(*configs.CryptoKey) types.CryptoKey
}
