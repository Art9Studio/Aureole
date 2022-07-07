package facebook

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"net/http"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"

	_ "embed"
	"errors"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Meta

// init initializes package by register pluginCreator
func init() {
	meta = core.AuthenticatorRepo.Register(rawMeta, Create)

}

type facebook struct {
	pluginAPI core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
	provider  *oauth2.Config
}

func (f *facebook) GetLoginMethod() string {
	return http.MethodGet
}

func (f *facebook) GetAuthRoute() *core.Route {
	return &core.Route{
		Path:   "/",
		Method: http.MethodGet,
		//
		Handler: f.GetLoginWrapper(),
	}
}

func Create(conf configs.PluginConfig) core.Authenticator {
	return &facebook{rawConf: conf}
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

	return nil
}

func (facebook) GetMetaData() core.Meta {
	return meta
}

func (f *facebook) GetLoginWrapper() core.AuthNLoginFunc {
	return func(c fiber.Ctx) (*core.AuthNResult, error) {
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

		return &core.AuthNResult{
			Cred: &core.Credential{
				Name:  core.Email,
				Value: email,
			},
			Identity: &core.Identity{
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

func (f *facebook) GetAppRoutes() []*core.Route {
	return []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    pathPrefix,
			Handler: getAuthCode(f),
		},
	}
}
