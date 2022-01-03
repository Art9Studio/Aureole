package facebook

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

const pluginID = "3030"

type facebook struct {
	pluginApi core.PluginAPI
	app       *core.App
	rawConf   *configs.Authn
	conf      *config
	provider  *oauth2.Config
}

func (f *facebook) Init(appName string, api core.PluginAPI) (err error) {
	f.pluginApi = api
	f.conf, err = initConfig(&f.rawConf.Config)
	if err != nil {
		return err
	}

	f.app, err = f.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	if err := initProvider(f); err != nil {
		return err
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

func (f *facebook) Login() plugins.AuthNLoginFunc {
	return func(c fiber.Ctx) (*identity.Credential, fiber.Map, error) {
		state := c.Query("state")
		if state != "state" {
			return nil, nil, errors.New("invalid state")
		}
		code := c.Query("code")
		if code == "" {
			return nil, nil, errors.New("code not found")
		}

		userData, err := getUserData(f, code)
		if err != nil {
			return nil, nil, err
		}

		if ok, err := f.app.Filter(convertUserData(userData), f.rawConf.Filter); err != nil {
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
				identity.SocialID:      userData["id"],
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

func initProvider(f *facebook) error {
	url, err := f.app.GetUrl()
	if err != nil {
		return err
	}

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
	f.pluginApi.AddAppRoutes(f.app.GetName(), routes)
}
