package email

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "4071"

type (
	email struct {
		pluginAPI     core.PluginAPI
		rawConf       *configs.Authn
		conf          *config
		sender        plugins.Sender
		magicLink     *url.URL
		tmpl, tmplExt string
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

	e.magicLink = createMagicLink(e)
	if err != nil {
		return err
	}

	tmpl, err := os.ReadFile(e.conf.TmplPath)
	if err != nil {
		e.tmpl = defaultTmpl
		e.tmplExt = "txt"
	} else {
		e.tmpl = string(tmpl)
		e.tmplExt = path.Ext(e.conf.TmplPath)
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

func (e *email) GetLoginHandler() (string, func() plugins.AuthNLoginFunc) {
	return http.MethodGet, e.login
}

func (e *email) login() plugins.AuthNLoginFunc {
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

func createMagicLink(e *email) *url.URL {
	u := e.pluginAPI.GetAppUrl()
	u.Path = path.Clean(u.Path + loginUrl)
	return &u
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
