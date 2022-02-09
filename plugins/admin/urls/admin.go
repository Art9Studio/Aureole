package urls

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/spec"
	"net/http"
)

const pluginID = "4892"

const getUrlsPath = "urls"

type admin struct {
	pluginApi core.PluginAPI
	rawConf   *configs.Admin
	swagger   struct {
		Paths       *spec.Paths
		Definitions spec.Definitions
	}
}

//go:embed docs/swagger.json
var swaggerJson []byte

func (a *admin) Init(api core.PluginAPI) (err error) {
	a.pluginApi = api

	err = json.Unmarshal(swaggerJson, &a.swagger)
	if err != nil {
		fmt.Printf("urls admin plugin: cannot marshal swagger docs: %v", err)
	}

	createRoutes(a)
	return nil
}

func (a *admin) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: a.rawConf.Name,
		ID:   pluginID,
	}
}

func (a *admin) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	return a.swagger.Paths, a.swagger.Definitions
}

func createRoutes(u *admin) {
	routes := []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    getUrlsPath,
			Handler: getUrls(u),
		},
	}
	u.pluginApi.AddProjectRoutes(routes)
}
