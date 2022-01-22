package google

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"context"
	"net/http"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

const pluginID = "1010"

type google struct {
	pluginAPI core.PluginAPI
	rawConf   *configs.Authn
	conf      *config
	provider  *oauth2.Config
}

func (g *google) Init(api core.PluginAPI) (err error) {
	g.pluginAPI = api
	g.conf, err = initConfig(&g.rawConf.Config)
	if err != nil {
		return err
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
	return func(c fiber.Ctx) (*plugins.AuthNResult, error) {
		// todo: save state and compare later #2
		state := c.Query("state")
		if state != "state" {
			return nil, errors.New("invalid state")
		}
		code := c.Query("code")
		if code == "" {
			return nil, errors.New("code not found")
		}

		jwtT, err := getJwt(g, code)
		if err != nil {
			return nil, errors.New("error while exchange")
		}

		email, ok := jwtT.Get("email")
		if !ok {
			return nil, errors.New("can't get 'email' from token")
		}
		/*socialId, ok := jwtT.Get("sub")
		if !ok {
			return nil, errors.New("can't get 'social_id' from token")
		}*/
		userData, err := jwtT.AsMap(context.Background())
		if err != nil {
			return nil, err
		}

		ok, err = g.pluginAPI.Filter(convertUserData(userData), g.rawConf.Filter)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("input data doesn't pass filters")
		}

		return &plugins.AuthNResult{
			Cred: &plugins.Credential{
				Name:  plugins.Email,
				Value: email.(string),
			},
			Identity: &plugins.Identity{
				Email:         email.(*string),
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

func initProvider(g *google) error {
	redirectUri := g.pluginAPI.GetAppUrl()
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
	g.pluginAPI.AddAppRoutes(routes)
}
