package vk

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
	meta = core.Repo.Register(rawMeta, Create)
}

type vk struct {
	pluginAPI core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
	provider  *oauth2.Config
}

func (v *vk) GetLoginMethod() string {
	return http.MethodGet
}

func Create(conf configs.PluginConfig) core.Authenticator {
	return &vk{rawConf: conf}
}

func (v *vk) Init(api core.PluginAPI) (err error) {
	v.pluginAPI = api
	v.conf, err = initConfig(&v.rawConf.Config)
	if err != nil {
		return err
	}

	if err := initProvider(v); err != nil {
		return err
	}

	return nil
}

func (vk) GetMetaData() core.Meta {
	return meta
}

func (v *vk) GetLoginWrapper() core.AuthNLoginFunc {
	return func(c fiber.Ctx) (*core.AuthNResult, error) {
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

		ok, err := v.pluginAPI.Filter(convertUserData(userData), v.conf.Filter)
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

func (v *vk) GetPaths() []*core.Route {
	return []*core.Route{
		{
			Method:  http.MethodGet,
			Path:    pathPrefix,
			Handler: getAuthCode(v),
		},
	}
}
