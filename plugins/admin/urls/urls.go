package urls

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"net/http"
)

const pluginID = "4892"

type urls struct {
	pluginApi core.PluginAPI
	rawConf   *configs.Admin
}

func (u *urls) Init(api core.PluginAPI) (err error) {
	u.pluginApi = api
	createRoutes(u)
	return nil
}

func (u *urls) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: u.rawConf.Name,
		ID:   pluginID,
	}
}

func createRoutes(u *urls) {
	routes := []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    getUrlsPath,
			Handler: getUrls(u),
		},
	}
	u.pluginApi.AddProjectRoutes(routes)
}
