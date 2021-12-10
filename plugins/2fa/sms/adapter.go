package sms

import (
	factor2 "aureole/internal/plugins/2fa"
)

// AdapterName is the internal name of the adapter
const AdapterName = "sms"

// init initializes package by register adapter
func init() {
	factor2.Repository.Register(AdapterName, smsAdapter{})
}

type smsAdapter struct {
}
