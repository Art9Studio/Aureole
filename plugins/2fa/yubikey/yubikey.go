package yubikey

import (
	"aureole/internal/configs"
	"aureole/internal/plugins/core"
	app "aureole/internal/state/interface"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

const PluginID = "2734"

type yubikey struct {
	pluginApi core.PluginAPI
	app       app.AppState
	rawConf   *configs.SecondFactor
	conf      *config
}

func (y *yubikey) Init(appName string, api core.PluginAPI) (err error) {
	y.pluginApi = api
	y.conf, err = initConfig(&y.rawConf.Config)
	if err != nil {
		return err
	}

	y.app, err = y.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	createRoutes(y)
	return nil
}

func (*yubikey) GetPluginID() string {
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

func createRoutes(*yubikey) {

}
