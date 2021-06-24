package vk

import (
	"aureole/internal/plugins/authn"
)

// AdapterName is the internal name of the adapter
const AdapterName = "vk"

// init initializes package by register adapter
func init() {
	authn.Repository.Register(AdapterName, vkAdapter{})
}

type vkAdapter struct {
}
