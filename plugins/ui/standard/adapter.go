package standard

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "standard"

// init initializes package by register adapter
func init() {
	plugins.UIRepo.Register(adapterName, adapter{})
}

type adapter struct {
}

func (adapter) Create(conf *configs.UI) plugins.UI {
	return &ui{rawConf: conf}
}
