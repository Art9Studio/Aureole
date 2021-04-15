package cryptokey

import (
	"aureole/configs"
	ctxTypes "aureole/context/types"
	"aureole/internal/plugins"
	"aureole/internal/plugins/cryptokey/types"
)

var Repository = plugins.InitRepository()

// Adapter defines methods for authentication plugins
type Adapter interface {
	// Create returns desired crypto key depends on the given config
	Create(*configs.CryptoKey) types.CryptoKey
}

func InitRepository(context *ctxTypes.ProjectCtx) {
	Repository.ProjectCtx = context
}
