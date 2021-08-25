package urls

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/admin"
	_interface "aureole/internal/router/interface"
	"github.com/mitchellh/mapstructure"
)

type urls struct {
	rawConf *configs.Admin
	conf    *config
}

func (u *urls) Init() (err error) {
	if u.conf, err = initConfig(&u.rawConf.Config); err != nil {
		return err
	}
	createRoutes(u)
	return nil
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
	admin.Repository.PluginApi.Router.AddProjectRoutes(routes)
}
