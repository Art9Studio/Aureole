package plugins

import (
	"aureole/internal/configs"
)

type (
	// RootPluginCreator defines methods for admin plugin
	RootPluginCreator interface {
		// Create returns desired admin plugin depends on the given config
		Create(admin *configs.PluginConfig) RootPlugin
	}

	RootPlugin interface {
		MetaDataGetter
	}
)
