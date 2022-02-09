package pwbased

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"aureole/plugins/authn/pwbased/pwhasher"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/structs"
	"github.com/go-openapi/spec"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "7157"

type (
	authn struct {
		pluginAPI core.PluginAPI
		rawConf   *configs.Authn
		conf      *config
		manager   plugins.IDManager
		pwHasher  pwhasher.PWHasher
		reset     struct {
			sender      plugins.Sender
			tmpl        string
			tmplExt     string
			confirmLink *url.URL
		}
		verify struct {
			sender      plugins.Sender
			tmpl        string
			tmplExt     string
			confirmLink *url.URL
		}
		swagger struct {
			Paths       *spec.Paths
			Definitions spec.Definitions
		}
	}

	credentialInput struct {
		Id       interface{} `json:"id,omitempty"`
		Email    string      `json:"email,omitempty"`
		Phone    string      `json:"phone,omitempty"`
		Username string      `json:"username,omitempty"`
		Password string      `json:"password"`
	}

	email struct {
		Email string `json:"email"`
	}

	linkType string
)

const (
	ResetLink  linkType = "reset"
	VerifyLink linkType = "verify"
)

//go:embed docs/swagger.json
var swaggerJson []byte

func (c *credentialInput) AsMap() map[string]interface{} {
	structsConf := structs.New(c)
	structsConf.TagName = "mapstructure"
	return structsConf.Map()
}

func (a *authn) Init(api core.PluginAPI) (err error) {
	a.pluginAPI = api
	a.conf, err = initConfig(&a.rawConf.Config)
	if err != nil {
		return err
	}

	a.manager, err = a.pluginAPI.GetIDManager()
	if err != nil {
		return err
	}
	err = a.manager.CheckFeaturesAvailable([]string{"OnUserAuthenticated", "Register", "GetData", "Update"})
	if err != nil {
		return fmt.Errorf("cannot check id manager features available: %v", err)
	}

	a.pwHasher, err = pwhasher.NewPWHasher(a.conf.MainHasher)
	if err != nil {
		return fmt.Errorf("cannot intit pwbased authn: %v", err)
	}

	if resetEnabled(a) {
		a.reset.sender, err = a.pluginAPI.GetSender(a.conf.Reset.Sender)
		if err != nil {
			return err
		}

		resetTmpl, err := os.ReadFile(a.conf.Reset.TmplPath)
		if err != nil {
			a.reset.tmpl = defaultResetTmpl
			a.reset.tmplExt = "txt"
		} else {
			a.reset.tmpl = string(resetTmpl)
			a.reset.tmplExt = path.Ext(a.conf.Reset.TmplPath)
		}

		a.reset.confirmLink = createConfirmLink(ResetLink, a)
		if err != nil {
			return err
		}
	}

	if verifyEnabled(a) {
		a.verify.sender, err = a.pluginAPI.GetSender(a.conf.Verify.Sender)
		if err != nil {
			return err
		}

		verifyTmpl, err := os.ReadFile(a.conf.Verify.TmplPath)
		if err != nil {
			a.verify.tmpl = defaultVerifyTmpl
			a.verify.tmplExt = "txt"
		} else {
			a.verify.tmpl = string(verifyTmpl)
			a.verify.tmplExt = path.Ext(a.conf.Verify.TmplPath)
		}

		a.verify.confirmLink = createConfirmLink(VerifyLink, a)
		if err != nil {
			return err
		}
	}

	err = json.Unmarshal(swaggerJson, &a.swagger)
	if err != nil {
		fmt.Printf("pwbased authn: cannot marshal swagger docs: %v", err)
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
		var input credentialInput
		if err := c.BodyParser(&input); err != nil {
			return nil, err
		}
		if input.Password == "" {
			return nil, errors.New("password required")
		}

		ident, err := plugins.NewIdentity(input.AsMap())
		if err != nil {
			return nil, err
		}
		cred, err := getCredential(&input)
		if err != nil {
			return nil, err
		}

		pw, err := a.manager.GetData(cred, adapterName, plugins.Password)
		if err != nil {
			return nil, err
		}

		isMatch, err := a.pwHasher.ComparePw(input.Password, pw.(string))
		if err != nil {
			return nil, err
		}

		if isMatch {
			return &plugins.AuthNResult{
				Cred:     cred,
				Identity: ident,
				Provider: adapterName,
			}, nil
		} else {
			return nil, errors.New("wrong password")
		}
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

func resetEnabled(p *authn) bool {
	return p.conf.Reset.Sender != "" && p.conf.Reset.TmplPath != ""
}

func verifyEnabled(p *authn) bool {
	return p.conf.Verify.Sender != "" && p.conf.Verify.TmplPath != ""
}

func createConfirmLink(linkType linkType, p *authn) *url.URL {
	u := p.pluginAPI.GetAppUrl()
	switch linkType {
	case ResetLink:
		u.Path = path.Clean(u.Path + resetConfirmUrl)
	case VerifyLink:
		u.Path = path.Clean(u.Path + verifyConfirmUrl)
	}
	return &u
}

func createRoutes(p *authn) {
	routes := []*core.Route{
		{
			Method:  http.MethodPost,
			Path:    pathPrefix + registerUrl,
			Handler: register(p),
		},
	}

	if resetEnabled(p) {
		resetRoutes := []*core.Route{
			{
				Method:  http.MethodPost,
				Path:    pathPrefix + resetUrl,
				Handler: Reset(p),
			},
			{
				Method:  http.MethodGet,
				Path:    pathPrefix + resetConfirmUrl,
				Handler: ResetConfirm(p),
			},
		}
		routes = append(routes, resetRoutes...)
	}

	if verifyEnabled(p) {
		verifRoutes := []*core.Route{
			{
				Method:  http.MethodPost,
				Path:    pathPrefix + verifyUrl,
				Handler: Verify(p),
			},
			{
				Method:  http.MethodGet,
				Path:    pathPrefix + verifyConfirmUrl,
				Handler: VerifyConfirm(p),
			},
		}
		routes = append(routes, verifRoutes...)
	}

	p.pluginAPI.AddAppRoutes(routes)
}
