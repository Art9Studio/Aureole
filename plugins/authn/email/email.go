package email

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "4071"

type (
	email struct {
		pluginApi core.PluginAPI
		app       *core.App
		rawConf   *configs.Authn
		conf      *config
		sender    plugins.Sender
		magicLink *url.URL
	}

	input struct {
		Email string `json:"email"`
	}
)

func (e *email) Init(appName string, api core.PluginAPI) (err error) {
	e.pluginApi = api
	e.conf, err = initConfig(&e.rawConf.Config)
	if err != nil {
		return err
	}

	e.app, err = e.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	e.sender, err = e.pluginApi.GetSender(e.conf.Sender)
	if err != nil {
		return fmt.Errorf("sender named '%s' is not declared", e.conf.Sender)
	}

	e.magicLink, err = createMagicLink(e)
	if err != nil {
		return err
	}

	createRoutes(e)
	return nil
}

func (*email) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		ID:   pluginID,
	}
}

func (*email) Login() plugins.AuthNLoginFunc {
	return func(c fiber.Ctx) (*plugins.AuthNResult, error) {
		rawToken := c.Query("token")
		if rawToken == "" {
			return nil, errors.New("token not found")
		}

		token, err := core.ParseJWT(rawToken)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return nil, errors.New("cannot get email from token")
		}
		if err := core.InvalidateJWT(token); err != nil {
			return nil, errors.New(err.Error())
		}

		return &plugins.AuthNResult{
			Cred: &plugins.Credential{
				Name:  plugins.Email,
				Value: email.(string),
			},
			Identity: &plugins.Identity{
				Email:         email.(*string),
				EmailVerified: true,
			},
			Provider: adapterName,
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

func createMagicLink(e *email) (*url.URL, error) {
	u, err := e.app.GetUrl()
	if err != nil {
		return nil, err
	}

	u.Path = path.Clean(u.Path + loginUrl)
	return &u, nil
}

func createRoutes(e *email) {
	routes := []*core.Route{
		{
			Method:  http.MethodPost,
			Path:    sendUrl,
			Handler: sendMagicLink(e),
		},
	}
	e.pluginApi.AddAppRoutes(e.app.GetName(), routes)
}
