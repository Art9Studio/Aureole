package sms

import (
	"aureole/internal/core"
)

// name is the internal name of the plugin
const name = "sms"

// init initializes package by register plugin
func init() {
	core.Repo.Register(name, smsPlugin{})
}

type smsPlugin struct {
}
