package google

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins"
	authnT "aureole/internal/plugins/authn/types"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/core"
	"aureole/internal/router"
	app "aureole/internal/state/interface"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"path"
)

const PluginID = "1010"

type google struct {
	pluginApi  core.PluginAPI
	app        app.AppState
	rawConf    *configs.Authn
	conf       *config
	manager    identity.ManagerI
	provider   *oauth2.Config
	authorizer authzTypes.Authorizer
}

func (g *google) Init(appName string, api core.PluginAPI) (err error) {
	g.pluginApi = api
	g.conf, err = initConfig(&g.rawConf.Config)
	if err != nil {
		return err
	}

	g.app, err = g.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	g.manager, err = g.app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared", appName)
	}

	g.authorizer, err = g.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", appName)
	}

	if err := initProvider(g); err != nil {
		return err
	}
	createRoutes(g)
	return nil
}

func (*google) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: AdapterName,
		ID:   PluginID,
	}
}

func (g *google) Login() authnT.AuthFunc {
	return func(c fiber.Ctx) (*identity.Credential, fiber.Map, error) {
		// todo: save state and compare later #2
		state := c.Query("state")
		if state != "state" {
			return nil, nil, errors.New("invalid state")
		}
		code := c.Query("code")
		if code == "" {
			return nil, nil, errors.New("code not found")
		}

		jwtT, err := getJwt(g, code)
		if err != nil {
			return nil, nil, errors.New("error while exchange")
		}

		email, ok := jwtT.Get("email")
		if !ok {
			return nil, nil, errors.New("can't get 'email' from token")
		}
		socialId, ok := jwtT.Get("sub")
		if !ok {
			return nil, nil, errors.New("can't get 'social_id' from token")
		}
		userData, err := jwtT.AsMap(context.Background())
		if err != nil {
			return nil, nil, err
		}

		if ok, err := g.app.Filter(convertUserData(userData), g.rawConf.Filter); err != nil {
			return nil, nil, err
		} else if !ok {
			return nil, nil, errors.New("input data doesn't pass filters")
		}

		return &identity.Credential{
				Name:  identity.Email,
				Value: email.(string),
			},
			fiber.Map{
				identity.Email:         email,
				identity.AuthnProvider: AdapterName,
				identity.SocialID:      socialId,
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

func initProvider(g *google) error {
	redirectUri, err := g.app.GetUrl()
	if err != nil {
		return err
	}

	redirectUri.Path = path.Clean(redirectUri.Path + g.conf.PathPrefix + g.conf.RedirectUri)
	g.provider = &oauth2.Config{
		ClientID:     g.conf.ClientId,
		ClientSecret: g.conf.ClientSecret,
		Endpoint:     endpoints.Google,
		RedirectURL:  redirectUri.String(),
		Scopes:       g.conf.Scopes,
	}
	return nil
}

func createRoutes(g *google) {
	routes := []*router.Route{
		{
			Method:  router.MethodGET,
			Path:    g.conf.PathPrefix,
			Handler: GetAuthCode(g),
		},
	}
	router.GetRouter().AddAppRoutes(g.app.GetName(), routes)
}
