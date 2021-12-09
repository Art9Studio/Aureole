package google

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/identity"
	"aureole/internal/plugins"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"net/http"
	"path"
)

const pluginID = "1010"

type google struct {
	pluginApi core.PluginAPI
	app       *core.App
	rawConf   *configs.Authn
	conf      *config
	provider  *oauth2.Config
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

	if err := initProvider(g); err != nil {
		return err
	}
	createRoutes(g)
	return nil
}

func (*google) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		ID:   pluginID,
	}
}

func (g *google) Login() plugins.AuthNLoginFunc {
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
				identity.AuthnProvider: adapterName,
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

	redirectUri.Path = path.Clean(redirectUri.Path + pathPrefix + redirectUrl)
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
	routes := []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    pathPrefix,
			Handler: getAuthCode(g),
		},
	}
	g.pluginApi.AddAppRoutes(g.app.GetName(), routes)
}
