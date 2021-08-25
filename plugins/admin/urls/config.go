package urls

import (
	"aureole/internal/configs"
	adminTypes "aureole/internal/plugins/admin/types"
)

type config struct {
	Path string `mapstructure:"path"`
}

// Create returns urls hasher with the given settings
func (a urlsAdapter) Create(conf *configs.Admin) adminTypes.Admin {
	return &urls{rawConf: conf}
}
