package yubikey

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "2734"

type yubikey struct {
	pluginApi core.PluginAPI
	app       *core.App
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
		Type: adapterName,
		Name: y.rawConf.Name,
		ID:   pluginID,
	}
}

func (y *yubikey) IsEnabled(cred *plugins.Credential, provider string) (bool, error) {
	enabled, id, err := y.pluginApi.Is2FAEnabled(cred, provider)
	if err != nil {
		return false, err
	}
	if !enabled {
		return false, nil
	}
	if id != pluginID {
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

func (*yubikey) Init2FA(_ *plugins.Credential, _ string, _ fiber.Ctx) (fiber.Map, error) {
	// TODO implement me
	panic("implement me")
}

func (*yubikey) Verify() plugins.MFAVerifyFunc {
	// TODO implement me
	panic("implement me")
}

func createRoutes(*yubikey) {

}
