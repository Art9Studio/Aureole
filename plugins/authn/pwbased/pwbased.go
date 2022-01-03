package pwbased

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/identity"
	"aureole/internal/plugins"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"net/url"
	"path"
)

const pluginID = "7157"

type (
	pwBased struct {
		pluginApi         core.PluginAPI
		app               *core.App
		rawConf           *configs.Authn
		conf              *config
		manager           identity.ManagerI
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

func (p *pwBased) Init(appName string, api core.PluginAPI) (err error) {
	p.pluginApi = api
	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	p.app, err = p.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	p.manager, err = p.app.GetIdentityManager()
	if err != nil {
		return fmt.Errorf("manager for app '%s' is not declared", appName)
	}
	if err := p.manager.CheckFeaturesAvailable([]string{"on_register", "get_data", "update"}); err != nil {
		return err
	}

	p.pwHasher, err = p.pluginApi.GetHasher(p.conf.MainHasher)
	if err != nil {
		return fmt.Errorf("hasher named '%s' is not declared", p.conf.MainHasher)
	}

	if resetEnabled(p) {
		p.resetSender, err = p.pluginApi.GetSender(p.conf.Reset.Sender)
		if err != nil {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Reset.Sender)
		}
		p.resetConfirmLink, err = createConfirmLink(ResetLink, p)
		if err != nil {
			return err
		}
	}

	if verifyEnabled(p) {
		p.verifySender, err = p.pluginApi.GetSender(p.conf.Verif.Sender)
		if err != nil {
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
	return func(c fiber.Ctx) (*identity.Credential, fiber.Map, error) {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return nil, nil, err
		}
		if input.Password == "" {
			return nil, nil, errors.New("password required")
		}

		i := &identity.Identity{
			ID:       input.Id,
			Email:    input.Email,
			Phone:    input.Phone,
			Username: input.Username,
		}
		cred, err := getCredential(i)
		if err != nil {
			return nil, nil, err
		}

		pw, err := p.manager.GetData(cred, adapterName, identity.Password)
		if err != nil {
			return nil, nil, err
		}

		isMatch, err := p.pwHasher.ComparePw(input.Password, pw.(string))
		if err != nil {
			return nil, nil, err
		}

		if isMatch {
			return cred, fiber.Map{
				cred.Name:              cred.Value,
				identity.AuthnProvider: adapterName,
			}, nil
		} else {
			return nil, nil, errors.New("wrong password")
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
	u, err := p.app.GetUrl()
	if err != nil {
		return nil, err
	}

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

	p.pluginApi.AddAppRoutes(p.app.GetName(), routes)
}
