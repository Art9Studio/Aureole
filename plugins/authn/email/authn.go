package email

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-openapi/spec"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "4071"

type (
	authn struct {
		pluginAPI     core.PluginAPI
		rawConf       *configs.Authn
		conf          *config
		sender        plugins.Sender
		magicLink     *url.URL
		tmpl, tmplExt string
		swagger       struct {
			Paths       *spec.Paths
			Definitions spec.Definitions
		}
	}

	input struct {
		Email string `json:"email"`
	}
)

//go:embed docs/swagger.json
var swaggerJson []byte

func (a *authn) Init(api core.PluginAPI) (err error) {
	a.pluginAPI = api
	a.conf, err = initConfig(&a.rawConf.Config)
	if err != nil {
		return err
	}

	a.sender, err = a.pluginAPI.GetSender(a.conf.Sender)
	if err != nil {
		return err
	}

	a.magicLink = createMagicLink(a)
	if err != nil {
		return err
	}

	err = json.Unmarshal(swaggerJson, &a.swagger)
	if err != nil {
		fmt.Printf("email authn: cannot marshal swagger docs: %v", err)
	}

	tmpl, err := os.ReadFile(a.conf.TmplPath)
	if err != nil {
		a.tmpl = defaultTmpl
		a.tmplExt = "txt"
	} else {
		a.tmpl = string(tmpl)
		a.tmplExt = path.Ext(a.conf.TmplPath)
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
		rawToken := c.Query("token")
		if rawToken == "" {
			return nil, errors.New("token not found")
		}

		var email string
		token, err := a.pluginAPI.ParseJWT(rawToken)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		err = a.pluginAPI.GetFromJWT(token, "email", &email)
		if err != nil {
			return nil, errors.New("cannot get email from token")
		}
		if err := a.pluginAPI.InvalidateJWT(token); err != nil {
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

func createMagicLink(e *authn) *url.URL {
	u := e.pluginAPI.GetAppUrl()
	u.Path = path.Clean(u.Path + loginUrl)
	return &u
}

func createRoutes(e *authn) {
	routes := []*core.Route{
		{
			Method:  http.MethodPost,
			Path:    sendUrl,
			Handler: sendMagicLink(e),
		},
	}
	e.pluginAPI.AddAppRoutes(routes)
}
