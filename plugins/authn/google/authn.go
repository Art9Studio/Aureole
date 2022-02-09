package google

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/spec"
	"net/http"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

const pluginID = "1010"

type authn struct {
	pluginAPI core.PluginAPI
	rawConf   *configs.Authn
	conf      *config
	provider  *oauth2.Config
	swagger   struct {
		Paths       *spec.Paths
		Definitions spec.Definitions
	}
}

//go:embed docs/swagger.json
var swaggerJson []byte

func (a *authn) Init(api core.PluginAPI) (err error) {
	a.pluginAPI = api
	a.conf, err = initConfig(&a.rawConf.Config)
	if err != nil {
		return err
	}
	if err := initProvider(a); err != nil {
		return err
	}

	err = json.Unmarshal(swaggerJson, &a.swagger)
	if err != nil {
		fmt.Printf("google authn: cannot marshal swagger docs: %v", err)
	}

	createRoutes(a)
	return nil
}

func (*authn) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		ID:   pluginID,
	}
}

func (a *authn) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	return a.swagger.Paths, a.swagger.Definitions
}

func (a *authn) LoginWrapper() plugins.AuthNLoginFunc {
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

		var email string
		jwtT, err := getJwt(a, code)
		if err != nil {
			return nil, errors.New("error while exchange")
		}
		err = a.pluginAPI.GetFromJWT(jwtT, "email", &email)
		if err != nil {
			return nil, errors.New("cannot get email from token")
		}
		/*socialId, ok := jwtT.Get("sub")
		if !ok {
			return nil, errors.New("can't get 'social_id' from token")
		}*/
		userData, err := jwtT.AsMap(context.Background())
		if err != nil {
			return nil, err
		}

		ok, err := a.pluginAPI.Filter(convertUserData(userData), a.rawConf.Filter)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("input data doesn't pass filters")
		}

		return &plugins.AuthNResult{
			Cred: &plugins.Credential{
				Name:  plugins.Email,
				Value: email,
			},
			Identity: &plugins.Identity{
				Email:         &email,
				EmailVerified: true,
				Additional: map[string]interface{}{
					"social_provider_data": map[string]interface{}{
						"plugin_id": pluginID, "payload": userData,
					},
				},
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

func initProvider(g *authn) error {
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

func createRoutes(g *authn) {
	routes := []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    pathPrefix,
			Handler: getAuthCode(g),
		},
	}
	g.pluginAPI.AddAppRoutes(routes)
}
