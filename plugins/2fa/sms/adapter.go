package sms

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "sms"

// init initializes package by register adapter
func init() {
	plugins.SecondFactorRepo.Register(adapterName, smsAdapter{})
}

type smsAdapter struct {
}
