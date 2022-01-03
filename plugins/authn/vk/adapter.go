package vk

import (
	"aureole/internal/plugins"
)

// adapterName is the internal name of the adapter
const adapterName = "vk"

// init initializes package by register adapter
func init() {
	plugins.AuthNRepo.Register(adapterName, vkAdapter{})
}

type vkAdapter struct {
}
