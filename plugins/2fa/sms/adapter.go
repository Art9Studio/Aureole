package sms

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "sms"

// init initializes package by register adapter
func init() {
	plugins.SecondFactorRepo.Register(adapterName, adapter{})
}

type adapter struct {
}

func (adapter) Create(conf *configs.SecondFactor) plugins.SecondFactor {
	return &mfa{rawConf: conf}
}
