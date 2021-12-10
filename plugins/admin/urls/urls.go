package urls

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/core"
	"aureole/internal/router/interface"
	"github.com/mitchellh/mapstructure"
)

const PluginID = "4892"

type urls struct {
	pluginApi core.PluginAPI
	rawConf   *configs.Admin
	conf      *config
}

func (u *urls) Init(api core.PluginAPI) (err error) {
	u.pluginApi = api
	if u.conf, err = initConfig(&u.rawConf.Config); err != nil {
		return err
	}
	u.conf.Path = "/admin/urls"
	createRoutes(u)
	return nil
}

func (*urls) GetPluginID() string {
	return PluginID
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()
	return adapterConf, nil
}

func createRoutes(u *urls) {
	routes := []*_interface.Route{
		{
			Method:  "GET",
			Path:    u.conf.Path,
			Handler: GetUrls(u),
		},
	}
	u.pluginApi.GetRouter().AddProjectRoutes(routes)
}
