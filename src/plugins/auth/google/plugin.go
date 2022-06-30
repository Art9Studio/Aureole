package google

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"context"
	_ "embed"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"net/http"
	"path"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Meta

// init initializes package by register pluginCreator
func init() {
	meta = core.Repo.Register(rawMeta, Create)
}

type google struct {
	pluginAPI core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
	provider  *oauth2.Config
}

func Create(conf configs.PluginConfig) core.Authenticator {
	return &google{rawConf: conf}
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

	return nil
}

func (google) GetMetaData() core.Meta {
	return meta
}

func (g *google) LoginWrapper() core.AuthNLoginFunc {
	return func(c fiber.Ctx) (*core.AuthNResult, error) {
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
		jwtT, err := getJwt(g, code)
		if err != nil {
			return nil, errors.New("error while exchange")
		}
		err = g.pluginAPI.GetFromJWT(jwtT, "email", &email)
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

		ok, err := g.pluginAPI.Filter(convertUserData(userData), g.conf.Filter)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("input data doesn't pass filters")
		}

		return &core.AuthNResult{
			Cred: &core.Credential{
				Name:  core.Email,
				Value: email,
			},
			Identity: &core.Identity{
				Email:         &email,
				EmailVerified: true,
				Additional: map[string]interface{}{
					"social_provider_data": map[string]interface{}{
						"plugin_name": meta.Name, "payload": userData,
					},
				},
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

func (g *google) GetPaths() []*core.Route {
	return []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    pathPrefix,
			Handler: getAuthCode(g),
		},
	}
}