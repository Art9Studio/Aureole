package standard

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/spec"
	"github.com/mitchellh/mapstructure"
	"net/http"
)

const pluginID = ""

type ui struct {
	pluginApi core.PluginAPI
	rawConf   *configs.UI
	conf      *config
	swagger   struct {
		Paths       *spec.Paths
		Definitions spec.Definitions
	}
}

var (
	//go:embed docs/swagger.json
	swaggerJson []byte
)

func (u *ui) Init(api core.PluginAPI) (err error) {
	u.pluginApi = api
	adapterConf := &config{}
	if err := mapstructure.Decode(u.rawConf.Config, adapterConf); err != nil {
		return err
	}
	adapterConf.setDefaults()
	u.conf = adapterConf

	err = json.Unmarshal(swaggerJson, &u.swagger)
	if err != nil {
		fmt.Printf("ui admin plugin: cannot marshal swagger docs: %v", err)
	}

	u.conf.setDefaults()
	createRoutes(u)
	return nil
}

func (u *ui) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: u.rawConf.Name,
		ID:   pluginID,
	}
}

func (u *ui) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	return u.swagger.Paths, u.swagger.Definitions
}

func createRoutes(u *ui) {
	routes := []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    getRedirectURL,
			Handler: getRedirect(u),
		},
		{
			Method:  http.MethodGet,
			Path:    getJWTStorageKeysURL,
			Handler: getJWTStorageKeys(u),
		},
	}
	u.pluginApi.AddAppRoutes(routes)
}
