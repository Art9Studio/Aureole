package facebook

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"fmt"
	"github.com/go-openapi/spec"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"net/http"
	"path"

	_ "embed"
	"encoding/json"
	"errors"
)

//go:embed swagger.json
var swaggerJson []byte

//go:embed meta.yaml
var rawMeta []byte

var meta plugins.Meta

// init initializes package by register pluginCreator
func init() {
	meta = plugins.Repo.Register(rawMeta, pluginCreator{})
}

type pluginCreator struct {
}

type facebook struct {
	pluginAPI core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
	provider  *oauth2.Config
	swagger   struct {
		Paths       *spec.Paths
		Definitions spec.Definitions
	}
}

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

func (facebook) GetMetaData() plugins.Meta {
	return meta
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

		ok, err := f.pluginAPI.Filter(convertUserData(userData), f.conf.Filter)
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
			Provider: "social_provider$" + meta.Name,
		}, nil
	}
}

func initConfig(conf *configs.RawConfig) (*config, error) {
	PluginConf := &config{}
	if err := mapstructure.Decode(conf, PluginConf); err != nil {
		return nil, err
	}
	PluginConf.setDefaults()
	return PluginConf, nil
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
