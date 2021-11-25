package facebook

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins"
	authnT "aureole/internal/plugins/authn/types"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/core"
	"aureole/internal/router"
	app "aureole/internal/state/interface"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"path"
)

const PluginID = "3030"

type facebook struct {
	pluginApi  core.PluginAPI
	app        app.AppState
	rawConf    *configs.Authn
	conf       *config
	manager    identity.ManagerI
	provider   *oauth2.Config
	authorizer authzTypes.Authorizer
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

	f.manager, err = f.app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared, persist layer is not available", appName)
	}

	f.authorizer, err = f.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", appName)
	}

	if err := initProvider(f); err != nil {
		return err
	}
	createRoutes(f)
	return nil
}

func (*facebook) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: AdapterName,
		ID:   PluginID,
	}
}

func (f *facebook) Login() authnT.AuthFunc {
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
				identity.AuthnProvider: AdapterName,
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
	redirectUri, err := f.app.GetUrl()
	if err != nil {
		return err
	}

	redirectUri.Path = path.Clean(redirectUri.Path + f.conf.PathPrefix + f.conf.RedirectUri)
	f.provider = &oauth2.Config{
		ClientID:     f.conf.ClientId,
		ClientSecret: f.conf.ClientSecret,
		Endpoint:     endpoints.Facebook,
		RedirectURL:  redirectUri.String(),
		Scopes:       f.conf.Scopes,
	}
	return nil
}

func createRoutes(f *facebook) {
	routes := []*router.Route{
		{
			Method:  router.MethodGET,
			Path:    f.conf.PathPrefix,
			Handler: GetAuthCode(f),
		},
	}
	router.GetRouter().AddAppRoutes(f.app.GetName(), routes)
}
