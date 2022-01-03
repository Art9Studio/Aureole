package vk

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/identity"
	"aureole/internal/plugins"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"net/http"
	"path"
)

const pluginID = "3888"

type vk struct {
	pluginApi  core.PluginAPI
	app        *core.App
	rawConf    *configs.Authn
	conf       *config
	manager    identity.ManagerI
	provider   *oauth2.Config
	authorizer plugins.Authorizer
}

func (v *vk) Init(appName string, api core.PluginAPI) (err error) {
	v.pluginApi = api
	v.conf, err = initConfig(&v.rawConf.Config)
	if err != nil {
		return err
	}

	v.app, err = v.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	v.manager, err = v.app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared", appName)
	}

	v.authorizer, err = v.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", appName)
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
	return func(c fiber.Ctx) (*identity.Credential, fiber.Map, error) {
		state := c.Query("state")
		if state != "state" {
			return nil, nil, errors.New("invalid state")
		}
		code := c.Query("code")
		if code == "" {
			return nil, nil, errors.New("code not found")
		}

		userData, err := getUserData(v, code)
		if err != nil {
			return nil, nil, err
		}

		if ok, err := v.app.Filter(convertUserData(userData), v.rawConf.Filter); err != nil {
			return nil, nil, err
		} else if !ok {
			return nil, nil, errors.New("input data doesn't pass filters")
		}

		return &identity.Credential{
				Name:  identity.Email,
				Value: userData["email"].(string),
			},
			fiber.Map{
				identity.Email:         userData["email"],
				identity.AuthnProvider: adapterName,
				identity.SocialID:      userData["user_id"],
				identity.UserData:      userData,
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
	redirectUri, err := v.app.GetUrl()
	if err != nil {
		return err
	}

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
	v.pluginApi.AddAppRoutes(v.app.GetName(), routes)
}
