package authenticator

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "google_authenticator"

// init initializes package by register adapter
func init() {
	plugins.SecondFactorRepo.Register(adapterName, gauthAdapter{})
}

type gauthAdapter struct {
}

func (gauthAdapter) Create(conf *configs.SecondFactor) plugins.SecondFactor {
	return &mfa{rawConf: conf}
}
