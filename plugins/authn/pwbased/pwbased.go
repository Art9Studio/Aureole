package pwbased

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

const pluginID = "7157"

type (
	pwBased struct {
		pluginAPI         core.PluginAPI
		rawConf           *configs.Authn
		conf              *config
		manager           plugins.IDManager
		pwHasher          plugins.PWHasher
		resetSender       plugins.Sender
		resetConfirmLink  *url.URL
		verifySender      plugins.Sender
		verifyConfirmLink *url.URL
	}

	input struct {
		Id       interface{} `json:"id"`
		Email    string      `json:"email"`
		Phone    string      `json:"phone"`
		Username string      `json:"username"`
		Password string      `json:"password"`
	}

	linkType string
)

const (
	ResetLink  linkType = "reset"
	VerifyLink linkType = "verify"
)

func (p *pwBased) Init(api core.PluginAPI) (err error) {
	p.pluginAPI = api
	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	var ok bool
	p.manager, ok = p.pluginAPI.GetIDManager()
	if !ok {
		return fmt.Errorf("manager for app '%s' is not declared", p.pluginAPI.GetAppName())
	}
	err = p.manager.CheckFeaturesAvailable([]string{"OnUserAuthenticated", "Register", "GetData", "Update"})
	if err != nil {
		return fmt.Errorf("cannot check id manager features available: %v", err)
	}

	p.pwHasher, ok = p.pluginAPI.GetHasher(p.conf.MainHasher)
	if !ok {
		return fmt.Errorf("hasher named '%s' is not declared", p.conf.MainHasher)
	}

	if resetEnabled(p) {
		p.resetSender, ok = p.pluginAPI.GetSender(p.conf.Reset.Sender)
		if !ok {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Reset.Sender)
		}
		p.resetConfirmLink, err = createConfirmLink(ResetLink, p)
		if err != nil {
			return err
		}
	}

	if verifyEnabled(p) {
		p.verifySender, ok = p.pluginAPI.GetSender(p.conf.Verif.Sender)
		if !ok {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Verif.Sender)
		}
		p.verifyConfirmLink, err = createConfirmLink(VerifyLink, p)
		if err != nil {
			return err
		}
	}

	createRoutes(p)
	return nil
}

func (*pwBased) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		ID:   pluginID,
	}
}

func (p *pwBased) Login() plugins.AuthNLoginFunc {
	return func(c fiber.Ctx) (*plugins.AuthNResult, error) {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return nil, err
		}
		if input.Password == "" {
			return nil, errors.New("password required")
		}

		ident := &plugins.Identity{
			ID:       input.Id,
			Email:    &input.Email,
			Phone:    &input.Phone,
			Username: &input.Username,
		}
		cred, err := getCredential(ident)
		if err != nil {
			return nil, err
		}

		pw, err := p.manager.GetData(cred, adapterName, plugins.Password)
		if err != nil {
			return nil, err
		}

		isMatch, err := p.pwHasher.ComparePw(input.Password, pw.(string))
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

func resetEnabled(p *pwBased) bool {
	return p.conf.Reset.Sender != "" && p.conf.Reset.Template != ""
}

func verifyEnabled(p *pwBased) bool {
	return p.conf.Verif.Sender != "" && p.conf.Verif.Template != ""
}

func createConfirmLink(linkType linkType, p *pwBased) (*url.URL, error) {
	u := p.pluginAPI.GetAppUrl()
	switch linkType {
	case ResetLink:
		u.Path = path.Clean(u.Path + resetConfirmUrl)
	case VerifyLink:
		u.Path = path.Clean(u.Path + verifyConfirmUrl)
	}

	return &u, nil
}

func createRoutes(p *pwBased) {
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
				Method:  http.MethodPost,
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
