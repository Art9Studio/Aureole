package core

import (
	"aureole/internal/collections"
	ctx "aureole/internal/context/interface"
	_interface "aureole/internal/router/interface"
)

type PluginApi struct {
	Project ctx.ProjectCtx
	Router  _interface.IRouter
}

var pluginApi PluginApi

func InitApi(ctx ctx.ProjectCtx, router _interface.IRouter) {
	pluginApi = PluginApi{Project: ctx, Router: router}
}

func (api *PluginApi) RegisterCollectionType(col *collections.CollectionType) {
	collections.Repository.Register(col)
}
