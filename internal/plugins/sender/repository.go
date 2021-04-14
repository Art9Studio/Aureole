package sender

import (
	"aureole/configs"
	ctxTypes "aureole/context/types"
	"aureole/internal/plugins"
	"aureole/internal/plugins/sender/types"
)

var Repository = plugins.InitRepository()

// Adapter defines methods for authentication plugins
type Adapter interface {
	// Create returns desired messenger depends on the given config
	Create(*configs.Sender) types.Sender
}

func InitRepository(context *ctxTypes.ProjectCtx) {
	Repository.ProjectCtx = context
}
