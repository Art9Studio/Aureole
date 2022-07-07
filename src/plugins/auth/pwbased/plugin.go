package pwbased

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/plugins/auth/pwbased/pwhasher"
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

var meta core.Meta

// init initializes package by register pluginCreator
func init() {
	meta = core.AuthenticatorRepo.Register(rawMeta, Create)

}

type (
	pwBased struct {
		pluginAPI core.PluginAPI
		rawConf   configs.PluginConfig
		conf      *config
		pwHasher  pwhasher.PWHasher
		reset     struct {
			sender      core.Sender
			tmpl        string
			tmplExt     string
			confirmLink *url.URL
		}
		verify struct {
			sender      core.Sender
			tmpl        string
			tmplExt     string
			confirmLink *url.URL
		}
	}

	credential struct {
		Id       interface{} `json:"id"`
		Email    string      `json:"email"`
		Phone    string      `json:"phone"`
		Username string      `json:"username"`
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

func (p *pwBased) GetLoginMethod() string {
	return http.MethodGet
}

func Create(conf configs.PluginConfig) core.Authenticator {
	return &pwBased{rawConf: conf}
}

func (p *pwBased) Init(api core.PluginAPI) (err error) {
	p.pluginAPI = api
	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	manager, ok := p.pluginAPI.GetIDManager()
	if !ok {
		return fmt.Errorf("manager for app '%s' is not declared", p.pluginAPI.GetAppName())
	}

	err = manager.CheckFeaturesAvailable([]string{"OnUserAuthenticated", "Register", "GetData", "Update"})
	if err != nil {
		return fmt.Errorf("cannot check id manager features available: %v", err)
	}

	p.pwHasher, err = pwhasher.NewPWHasher(p.conf.MainHasher)
	if err != nil {
		return fmt.Errorf("cannot intit pwbased authn: %v", err)
	}

	if resetEnabled(p) {
		p.reset.sender, ok = p.pluginAPI.GetSender(p.conf.Reset.Sender)
		if !ok {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Reset.Sender)
		}

		resetTmpl, err := os.ReadFile(p.conf.Reset.TmplPath)
		if err != nil {
			p.reset.tmpl = defaultResetTmpl
			p.reset.tmplExt = "txt"
		} else {
			p.reset.tmpl = string(resetTmpl)
			p.reset.tmplExt = path.Ext(p.conf.Reset.TmplPath)
		}

		p.reset.confirmLink = createConfirmLink(ResetLink, p)
		if err != nil {
			return err
		}
	}

	if verifyEnabled(p) {
		p.verify.sender, ok = p.pluginAPI.GetSender(p.conf.Verify.Sender)
		if !ok {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Verify.Sender)
		}

		verifyTmpl, err := os.ReadFile(p.conf.Verify.TmplPath)
		if err != nil {
			p.verify.tmpl = defaultVerifyTmpl
			p.verify.tmplExt = "txt"
		} else {
			p.verify.tmpl = string(verifyTmpl)
			p.verify.tmplExt = path.Ext(p.conf.Verify.TmplPath)
		}

		p.verify.confirmLink = createConfirmLink(VerifyLink, p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pwBased) GetMetaData() core.Meta {
	return meta
}

func (p *pwBased) GetLoginWrapper() core.AuthNLoginFunc {
	return func(c fiber.Ctx) (*core.AuthNResult, error) {
		var input *credential
		if err := c.BodyParser(input); err != nil {
			return nil, err
		}
		if input.Password == "" {
			return nil, errors.New("password required")
		}

		ident := &core.Identity{
			ID:       input.Id,
			Email:    &input.Email,
			Phone:    &input.Phone,
			Username: &input.Username,
		}
		cred, err := getCredential(ident)
		if err != nil {
			return nil, err
		}

		manager, ok := p.pluginAPI.GetIDManager()
		if !ok {
			return nil, fmt.Errorf("id manager for app '%s' is required but not declared", p.pluginAPI.GetAppName())
		}

		pw, err := manager.GetData(cred, meta.Name, core.Password)
		if err != nil {
			return nil, err
		}

		isMatch, err := p.pwHasher.ComparePw(input.Password, pw.(string))
		if err != nil {
			return nil, err
		}

		if isMatch {
			return &core.AuthNResult{
				Cred:     cred,
				Identity: ident,
				Provider: meta.Name,
			}, nil
		} else {
			return nil, errors.New("wrong password")
		}
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

func resetEnabled(p *pwBased) bool {
	return p.conf.Reset.Sender != "" && p.conf.Reset.TmplPath != ""
}

func verifyEnabled(p *pwBased) bool {
	return p.conf.Verify.Sender != "" && p.conf.Verify.TmplPath != ""
}

func createConfirmLink(linkType linkType, p *pwBased) *url.URL {
	u := p.pluginAPI.GetAppUrl()
	switch linkType {
	case ResetLink:
		u.Path = path.Clean(u.Path + resetConfirmUrl)
	case VerifyLink:
		u.Path = path.Clean(u.Path + verifyConfirmUrl)
	}
	return &u
}

func (p *pwBased) GetAppRoutes() []*core.Route {
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
	return routes
}
