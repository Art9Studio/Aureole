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
		pluginAPI core.PluginAPI
		rawConf   *configs.Authn
		conf      *config
		sender    plugins.Sender
		magicLink *url.URL
	}

	input struct {
		Email string `json:"email"`
	}
)

func (e *email) Init(api core.PluginAPI) (err error) {
	e.pluginAPI = api
	e.conf, err = initConfig(&e.rawConf.Config)
	if err != nil {
		return err
	}

	var ok bool
	e.sender, ok = e.pluginAPI.GetSender(e.conf.Sender)
	if !ok {
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

func (e *email) Login() plugins.AuthNLoginFunc {
	return func(c fiber.Ctx) (*plugins.AuthNResult, error) {
		rawToken := c.Query("token")
		if rawToken == "" {
			return nil, errors.New("token not found")
		}

		var email string
		token, err := e.pluginAPI.ParseJWT(rawToken)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		err = e.pluginAPI.GetFromJWT(token, "email", &email)
		if err != nil {
			return nil, errors.New("cannot get email from token")
		}
		if err := e.pluginAPI.InvalidateJWT(token); err != nil {
			return nil, errors.New(err.Error())
		}

		return &plugins.AuthNResult{
			Cred: &plugins.Credential{
				Name:  plugins.Email,
				Value: email,
			},
			Identity: &plugins.Identity{
				Email:         &email,
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
	u := e.pluginAPI.GetAppUrl()
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
	e.pluginAPI.AddAppRoutes(routes)
}
