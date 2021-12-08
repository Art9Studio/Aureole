package core

import (
	_interface "aureole/internal/router/interface"
	state "aureole/internal/state/interface"
)

type PluginApi struct {
	Project state.ProjectState
	Router  _interface.IRouter
}

var pluginApi PluginApi

func InitApi(p state.ProjectState, router _interface.IRouter) {
	pluginApi = PluginApi{Project: p, Router: router}
}
