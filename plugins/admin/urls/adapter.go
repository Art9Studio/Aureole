package urls

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "urls"

// init initializes package by register adapter
func init() {
	plugins.AdminRepo.Register(adapterName, adapter{})
}

// adapter represents adapter for argon2 pwhasher algorithm
type adapter struct {
}

// Create returns urls hasher with the given settings
func (adapter) Create(conf *configs.Admin) plugins.Admin {
	return &admin{rawConf: conf}
}
