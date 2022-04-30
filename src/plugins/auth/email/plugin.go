package email

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"fmt"
	"github.com/go-openapi/spec"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"net/url"
	"os"
	"path"

	_ "embed"
	"encoding/json"
	"errors"
)

//go:embed swagger.json
var swaggerJson []byte

//go:embed meta.yaml
var rawMeta []byte

var meta plugins.Meta

// init initializes package by register pluginCreator
func init() {
	meta = plugins.Repo.Register(rawMeta, pluginCreator{})
}

// pluginCreator represents plugin for password based authentication
type pluginCreator struct {
}

type (
	email struct {
		pluginAPI     core.PluginAPI
		rawConf       configs.PluginConfig
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

	err = json.Unmarshal(swaggerJson, &e.swagger)
	if err != nil {
		fmt.Printf("email authn: cannot marshal swagger docs: %v", err)
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

func (email) GetMetaData() plugins.Meta {
	return meta
}

func (e *email) GetHandlersSpec() (*spec.Paths, spec.Definitions) {
	return e.swagger.Paths, e.swagger.Definitions
}

func (e *email) LoginWrapper() plugins.AuthNLoginFunc {
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
			Provider: meta.Name,
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
