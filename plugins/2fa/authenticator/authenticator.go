package authenticator

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/core"
	"aureole/internal/router/interface"
	app "aureole/internal/state/interface"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

const PluginID = "1799"

type (
	gauth struct {
		pluginApi core.PluginAPI
		app       app.AppState
		rawConf   *configs.SecondFactor
		conf      *config
		manager   identity.ManagerI
	}

	input struct {
		Token string `json:"token"`
		Otp   string `json:"otp"`
	}
)

func (g *gauth) Init(appName string, api core.PluginAPI) (err error) {
	g.pluginApi = api
	g.conf, err = initConfig(&g.rawConf.Config)
	if err != nil {
		return err
	}

	g.app, err = g.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	createRoutes(g)
	return nil
}

func (*gauth) GetPluginID() string {
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

func createRoutes(g *gauth) {
	routes := []*_interface.Route{
		{
			Method:  "POST",
			Path:    g.conf.PathPrefix + g.conf.GetQRUrl,
			Handler: GetQR(g),
		},
		{
			Method:  "POST",
			Path:    g.conf.PathPrefix + g.conf.VerifyUrl,
			Handler: VerifyOTP(g),
		},
		{
			Method:  "POST",
			Path:    g.conf.PathPrefix + g.conf.GetScratchesUrl,
			Handler: GetScratchCodes(g),
		},
	}
	g.pluginApi.GetRouter().AddAppRoutes(g.app.GetName(), routes)
}
