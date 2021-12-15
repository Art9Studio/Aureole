package vk

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

const pluginID = "3888"

type vk struct {
	pluginAPI  core.PluginAPI
	rawConf    *configs.Authn
	conf       *config
	manager    plugins.IDManager
	provider   *oauth2.Config
	authorizer plugins.Authorizer
}

func (v *vk) Init(api core.PluginAPI) (err error) {
	v.pluginAPI = api
	v.conf, err = initConfig(&v.rawConf.Config)
	if err != nil {
		return err
	}

	var ok bool
	v.manager, ok = v.pluginAPI.GetIDManager()
	if !ok {
		fmt.Printf("manager for app '%s' is not declared", v.pluginAPI.GetAppName())
	}

	v.authorizer, ok = v.pluginAPI.GetAuthorizer()
	if !ok {
		return fmt.Errorf("authorizer named for app '%s' is not declared", v.pluginAPI.GetAppName())
	}

	if err := initProvider(v); err != nil {
		return err
	}
	createRoutes(v)
	return nil
}

func (*vk) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		ID:   pluginID,
	}
}

func (v *vk) Login() plugins.AuthNLoginFunc {
	return func(c fiber.Ctx) (*plugins.AuthNResult, error) {
		state := c.Query("state")
		if state != "state" {
			return nil, errors.New("invalid state")
		}
		code := c.Query("code")
		if code == "" {
			return nil, errors.New("code not found")
		}

		userData, err := getUserData(v, code)
		if err != nil {
			return nil, err
		}

		ok, err := v.pluginAPI.Filter(convertUserData(userData), v.rawConf.Filter)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("input data doesn't pass filters")
		}

		return &plugins.AuthNResult{
			Cred: &plugins.Credential{
				Name:  plugins.Email,
				Value: userData["email"].(string),
			},
			Identity: &plugins.Identity{
				Email:         userData["email"].(*string),
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

func initProvider(v *vk) error {
	redirectUri := v.pluginAPI.GetAppUrl()
	redirectUri.Path = path.Clean(redirectUri.Path + pathPrefix + redirectUrl)
	v.provider = &oauth2.Config{
		ClientID:     v.conf.ClientId,
		ClientSecret: v.conf.ClientSecret,
		Endpoint:     endpoints.Vk,
		RedirectURL:  redirectUri.String(),
		Scopes:       v.conf.Scopes,
	}
	return nil
}

func createRoutes(v *vk) {
	routes := []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    pathPrefix,
			Handler: getAuthCode(v),
		},
	}
	v.pluginAPI.AddAppRoutes(routes)
}
