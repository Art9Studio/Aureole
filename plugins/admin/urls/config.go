package urls

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

const getUrlsPath = "urls"

// Create returns urls hasher with the given settings
func (urlsAdapter) Create(conf *configs.Admin) plugins.Admin {
	return &urls{rawConf: conf}
}
