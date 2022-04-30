package urls

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"fmt"
	"github.com/go-openapi/spec"
	"net/http"

	_ "embed"
	"encoding/json"
)

//go:embed swagger.json
var swaggerJson []byte

//go:embed meta.yaml
var rawMeta []byte

var meta plugins.Meta

// init initializes package by register pluginCreator
func init() {
	meta = plugins.Repo.Register(rawMeta, pluginCreator{})
}

type pluginCreator struct {
}

type urls struct {
	pluginApi core.PluginAPI
	rawConf   configs.PluginConfig
	swagger   struct {
		Paths       *spec.Paths
		Definitions spec.Definitions
	}
}

func (u *urls) Init(api core.PluginAPI) (err error) {
	u.pluginApi = api

	err = json.Unmarshal(swaggerJson, &u.swagger)
	if err != nil {
		fmt.Printf("urls admin plugin: cannot marshal swagger docs: %v", err)
	}

	createRoutes(u)
	return nil
}

func (u urls) GetMetaData() plugins.Meta {
	return meta
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
