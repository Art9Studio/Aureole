package yubikey

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/spec"
	"github.com/gofiber/fiber/v2"

	"github.com/mitchellh/mapstructure"
)

const pluginID = "2734"

type yubikey struct {
	pluginAPI core.PluginAPI
	rawConf   *configs.SecondFactor
	conf      *config
}

//go:embed swagger.json
var swaggerJson []byte

func (y *yubikey) Init(api core.PluginAPI) (err error) {
	y.pluginAPI = api
	y.conf, err = initConfig(&y.rawConf.Config)
	if err != nil {
		return err
	}

	_, ok := y.pluginAPI.GetIDManager()
	if !ok {
		return fmt.Errorf("manager for app '%s' is not declared", y.pluginAPI.GetAppName())
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

func (y *yubikey) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	specs := struct {
		Paths       *spec.Paths
		Definitions spec.Definitions
	}{}
	err := json.Unmarshal(swaggerJson, &specs)
	if err != nil {
		return nil, nil
	}
	return specs.Paths, specs.Definitions
}

func (y *yubikey) IsEnabled(cred *plugins.Credential) (bool, error) {
	return y.pluginAPI.Is2FAEnabled(cred, pluginID)
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()
	return adapterConf, nil
}

func (*yubikey) Init2FA() plugins.MFAInitFunc {
	return func(c fiber.Ctx) (fiber.Map, error) {
		// TODO implement me
		panic("implement me")
	}
}

func (*yubikey) Verify() plugins.MFAVerifyFunc {
	// TODO implement me
	panic("implement me")
}

func createRoutes(*yubikey) {

}
