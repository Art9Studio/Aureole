package sms

import (
	"aureole/internal/plugins"
)

// name is the internal name of the plugin
const name = "sms"

// init initializes package by register plugin
func init() {
	plugins.Repo.Register(name, smsPlugin{})
}

type smsPlugin struct {
}
