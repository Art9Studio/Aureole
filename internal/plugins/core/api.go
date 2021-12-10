package core

import (
	_interface "aureole/internal/router/interface"
	state "aureole/internal/state/interface"
)

type PluginApi struct {
	Project state.PluginsState
	Router  _interface.IRouter
}

var pluginApi PluginApi

func InitApi(p state.PluginsState, router _interface.IRouter) {
	pluginApi = PluginApi{Project: p, Router: router}
}
