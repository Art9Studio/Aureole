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

type urls struct {
	pluginApi core.PluginAPI
	rawConf   *configs.Admin
	swagger   struct {
		Paths       *spec.Paths
		Definitions spec.Definitions
	}
}

//go:embed swagger.json
var swaggerJson []byte

func (u *urls) Init(api core.PluginAPI) (err error) {
	u.pluginApi = api

	err = json.Unmarshal(swaggerJson, &u.swagger)
	if err != nil {
		fmt.Printf("urls admin plugin: cannot marshal swagger docs: %v", err)
	}

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

func (u *urls) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	return u.swagger.Paths, u.swagger.Definitions
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
