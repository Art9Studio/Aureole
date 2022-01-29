package facebook

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-openapi/spec"
	"net/http"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

const pluginID = "3030"

type facebook struct {
	pluginAPI core.PluginAPI
	rawConf   *configs.Authn
	conf      *config
	provider  *oauth2.Config
	swagger   struct {
		Paths       *spec.Paths
		Definitions spec.Definitions
	}
}

//go:embed swagger.json
var swaggerJson []byte

func (f *facebook) Init(api core.PluginAPI) (err error) {
	f.pluginAPI = api
	f.conf, err = initConfig(&f.rawConf.Config)
	if err != nil {
		return err
	}
	if err := initProvider(f); err != nil {
		return err
	}

	err = json.Unmarshal(swaggerJson, &f.swagger)
	if err != nil {
		fmt.Printf("facebook authn: cannot marshal swagger docs: %v", err)
	}

	createRoutes(f)
	return nil
}

func (*facebook) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		ID:   pluginID,
	}
}

func (f *facebook) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	return f.swagger.Paths, f.swagger.Definitions
}

func (f *facebook) LoginWrapper() plugins.AuthNLoginFunc {
	return func(c fiber.Ctx) (*plugins.AuthNResult, error) {
		state := c.Query("state")
		if state != "state" {
			return nil, errors.New("invalid state")
		}
		code := c.Query("code")
		if code == "" {
			return nil, errors.New("code not found")
		}

		userData, err := getUserData(f, code)
		if err != nil {
			return nil, err
		}

		ok, err := f.pluginAPI.Filter(convertUserData(userData), f.rawConf.Filter)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("input data doesn't pass filters")
		}
		email := userData["email"].(string)

		return &plugins.AuthNResult{
			Cred: &plugins.Credential{
				Name:  plugins.Email,
				Value: email,
			},
			Identity: &plugins.Identity{
				Email:         &email,
				EmailVerified: true,
				Additional:    map[string]interface{}{"social_provider_data": userData},
			},
			Provider: "social_provider$" + adapterName,
		}, nil
	}
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()
	return adapterConf, nil
}

func initProvider(f *facebook) error {
	url := f.pluginAPI.GetAppUrl()
	url.Path = path.Clean(url.Path + pathPrefix + redirectUrl)
	f.provider = &oauth2.Config{
		ClientID:     f.conf.ClientId,
		ClientSecret: f.conf.ClientSecret,
		Endpoint:     endpoints.Facebook,
		RedirectURL:  url.String(),
		Scopes:       f.conf.Scopes,
	}
	return nil
}

func createRoutes(f *facebook) {
	routes := []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    pathPrefix,
			Handler: getAuthCode(f),
		},
	}
	f.pluginAPI.AddAppRoutes(routes)
}
