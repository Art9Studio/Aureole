package yubikey

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	_ "embed"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

// init initializes package by register pluginCreator
func init() {
	meta = core.MFARepo.Register(rawMeta, Create)
}

type yubikey struct {
	pluginAPI core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
}

func (y *yubikey) GetCustomAppRoutes() []*core.Route {
	//TODO implement me
	//panic("implement me")
	return nil
}

func (y *yubikey) GetOAS3AuthRequestBody() *openapi3.RequestBody {
	//TODO implement me
	//panic("implement me")
	return nil
}

func (y *yubikey) GetOAS3AuthParameters() openapi3.Parameters {
	//TODO implement me
	//panic("implement me")
	return nil
}

func Create(conf configs.PluginConfig) core.MFA {
	return &yubikey{rawConf: conf}
}

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

	return nil
}

func (y yubikey) GetMetadata() core.Metadata {
	return meta
}

//func (y *yubikey) GetPaths() *openapi3.Paths {
//	specs := struct {
//		Paths       *openapi3.Paths
//		Definitions openapi3.Definition
//	}{}
//	err := json.Unmarshal(swaggerJson, &specs)
//	if err != nil {
//		return nil, nil
//	}
//	return specs.Paths, specs.Definitions
//}

func (y *yubikey) IsEnabled(cred *core.Credential) (bool, error) {
	return y.pluginAPI.IsMFAEnabled(cred)
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	PluginConf := &config{}
	if err := mapstructure.Decode(rawConf, PluginConf); err != nil {
		return nil, err
	}
	PluginConf.setDefaults()
	return PluginConf, nil
}

func (*yubikey) InitMFA() core.MFAInitFunc {
	return func(c fiber.Ctx) (core.MFAResMap, error) {
		// TODO implement me
		//panic("implement me")
		return nil, nil
	}
}

func (*yubikey) Verify() core.MFAVerifyFunc {
	// TODO implement me
	//panic("implement me")
	return nil
}

func GetPaths() []*core.Route {
	return nil
}
