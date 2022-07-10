package email

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"

	_ "embed"
	"errors"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

// init initializes package by register pluginCreator
func init() {
	meta = core.AuthenticatorRepo.Register(rawMeta, Create)
}

type (
	email struct {
		pluginAPI     core.PluginAPI
		rawConf       configs.PluginConfig
		conf          *config
		sender        core.Sender
		magicLink     *url.URL
		tmpl, tmplExt string
	}

	input struct {
		Email string `json:"email"`
	}
)

func (e *email) GetAuthHTTPMethod() string {
	return http.MethodGet
}

func Create(conf configs.PluginConfig) core.Authenticator {
	return &email{rawConf: conf}
}

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

	return nil
}

func (email) GetMetadata() core.Metadata {
	return meta
}

func (e *email) GetAuthHandler() core.AuthHandlerFunc {
	return func(c fiber.Ctx) (*core.AuthResult, error) {
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

		return &core.AuthResult{
			Cred: &core.Credential{
				Name:  core.Email,
				Value: email,
			},
			Identity: &core.Identity{
				Email:         &email,
				EmailVerified: true,
			},
			Provider: meta.ShortName,
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

func (e *email) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{
		{
			Method:  http.MethodPost,
			Path:    sendUrl,
			Handler: sendMagicLink(e),
		},
	}
}
