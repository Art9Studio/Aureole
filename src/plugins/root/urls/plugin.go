package urls

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"net/http"

	_ "embed"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Meta

// init initializes package by register pluginCreator
func init() {
	meta = core.RootRepo.Register(rawMeta, Create)
}

type urls struct {
	pluginApi core.PluginAPI
	rawConf   configs.PluginConfig
}

// Create returns urls hasher with the given settings
func Create(conf configs.PluginConfig) core.RootPlugin {
	return &urls{rawConf: conf}
}

func (u *urls) Init(api core.PluginAPI) (err error) {
	u.pluginApi = api

	return nil
}

func (u urls) GetMetaData() core.Meta {
	return meta
}

func (u *urls) GetAppRoutes() []*core.Route {
	return []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    getUrlsPath,
			Handler: getUrls(u),
		},
	}
}
