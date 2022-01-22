package yubikey

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"github.com/gofiber/fiber/v2"

	"github.com/mitchellh/mapstructure"
)

const pluginID = "2734"

type yubikey struct {
	pluginApi core.PluginAPI
	rawConf   *configs.SecondFactor
	conf      *config
}

func (y *yubikey) Init(appName string, api core.PluginAPI) (err error) {
	y.pluginApi = api
	y.conf, err = initConfig(&y.rawConf.Config)
	if err != nil {
		return err
	}
	createRoutes(y)
	return nil
}

func (y yubikey) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: y.rawConf.Name,
		ID:   pluginID,
	}
}

func (y yubikey) IsEnabled(cred *plugins.Credential) (bool, error) {
	return y.pluginApi.Is2FAEnabled(cred, pluginID)
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()
	return adapterConf, nil
}

func (yubikey) Init2FA() plugins.MFAInitFunc {
	return func(c fiber.Ctx) (fiber.Map, error) {
		// TODO implement me
		panic("implement me")
	}
}

func (yubikey) Verify() plugins.MFAVerifyFunc {
	// TODO implement me
	panic("implement me")
}

func createRoutes(*yubikey) {

}
