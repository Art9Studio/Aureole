package yubikey

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins"
	mfaT "aureole/internal/plugins/2fa/types"
	"aureole/internal/plugins/core"
	app "aureole/internal/state/interface"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
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

func (y *yubikey) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: AdapterName,
		Name: y.rawConf.Name,
		ID:   PluginID,
	}
}

func (y *yubikey) IsEnabled(cred *identity.Credential, provider string) (bool, error) {
	enabled, id, err := y.pluginApi.Is2FAEnabled(cred, provider)
	if err != nil {
		return false, err
	}
	if !enabled {
		return false, nil
	}
	if id != PluginID {
		return false, errors.New("another 2FA is enabled")
	}
	return true, nil
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()
	return adapterConf, nil
}

func (*yubikey) Init2FA(_ *identity.Credential, _ string, _ fiber.Ctx) (fiber.Map, error) {
	// TODO implement me
	panic("implement me")
}

func (*yubikey) Verify() mfaT.MFAFunc {
	// TODO implement me
	panic("implement me")
}

func createRoutes(*yubikey) {

}
