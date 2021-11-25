package pwbased

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins"
	authnT "aureole/internal/plugins/authn/types"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/core"
	"aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	"aureole/internal/router"
	app "aureole/internal/state/interface"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"net/url"
	"path"
)

const PluginID = "7157"

type (
	pwBased struct {
		pluginApi  core.PluginAPI
		app        app.AppState
		rawConf    *configs.Authn
		conf       *config
		manager    identity.ManagerI
		pwHasher   types.PwHasher
		authorizer authzTypes.Authorizer
		reset      *reset
		verif      *verification
	}

	reset struct {
		sender      senderTypes.Sender
		confirmLink *url.URL
	}

	verification struct {
		sender      senderTypes.Sender
		confirmLink *url.URL
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

	p.authorizer, err = p.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", appName)
	}

	if pwResetEnable(p) {
		p.reset = &reset{}
		p.reset.sender, err = p.pluginApi.GetSender(p.conf.Reset.Sender)
		if err != nil {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Reset.Sender)
		}

		p.reset.confirmLink, err = createConfirmLink(ResetLink, p)
		if err != nil {
			return err
		}
	}

	if verifEnable(p) {
		p.verif = &verification{}
		p.verif.sender, err = p.pluginApi.GetSender(p.conf.Verif.Sender)
		if err != nil {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Verif.Sender)
		}

		p.verif.confirmLink, err = createConfirmLink(VerifyLink, p)
		if err != nil {
			return err
		}
	}

	createRoutes(p)
	return nil
}

func (*pwBased) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: AdapterName,
		ID:   PluginID,
	}
}

func (p *pwBased) Login() authnT.AuthFunc {
	return func(c fiber.Ctx) (*identity.Credential, fiber.Map, error) {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return nil, nil, err
		}
		if input.Password == "" {
			return nil, nil, errors.New("password required")
		}

		i := &identity.Identity{
			Id:       input.Id,
			Email:    input.Email,
			Phone:    input.Phone,
			Username: input.Username,
		}
		cred, err := getCredential(i)
		if err != nil {
			return nil, nil, err
		}

		pw, err := p.manager.GetData(cred, AdapterName, identity.Password)
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
				identity.AuthnProvider: AdapterName,
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

func pwResetEnable(p *pwBased) bool {
	return p.conf.Reset.Sender != "" && p.conf.Reset.Template != ""
}

func verifEnable(p *pwBased) bool {
	return p.conf.Verif.Sender != "" && p.conf.Verif.Template != ""
}

func createConfirmLink(linkType linkType, p *pwBased) (*url.URL, error) {
	u, err := p.app.GetUrl()
	if err != nil {
		return nil, err
	}

	switch linkType {
	case ResetLink:
		u.Path = path.Clean(u.Path + p.conf.PathPrefix + p.conf.Reset.ConfirmUrl)
	case VerifyLink:
		u.Path = path.Clean(u.Path + p.conf.PathPrefix + p.conf.Verif.ConfirmUrl)
	}

	return &u, nil
}

func createRoutes(p *pwBased) {
	routes := []*router.Route{
		{
			Method:  router.MethodPOST,
			Path:    p.conf.PathPrefix + p.conf.Register.Path,
			Handler: Register(p),
		},
	}

	if pwResetEnable(p) {
		resetRoutes := []*router.Route{
			{
				Method:  router.MethodPOST,
				Path:    p.conf.PathPrefix + p.conf.Reset.Path,
				Handler: Reset(p),
			},
			{
				Method:  router.MethodPOST,
				Path:    p.conf.PathPrefix + p.conf.Reset.ConfirmUrl,
				Handler: ResetConfirm(p),
			},
		}
		routes = append(routes, resetRoutes...)
	}

	if verifEnable(p) {
		verifRoutes := []*router.Route{
			{
				Method:  router.MethodPOST,
				Path:    p.conf.PathPrefix + p.conf.Verif.Path,
				Handler: Verify(p),
			},
			{
				Method:  router.MethodGET,
				Path:    p.conf.PathPrefix + p.conf.Verif.ConfirmUrl,
				Handler: VerifyConfirm(p),
			},
		}
		routes = append(routes, verifRoutes...)
	}

	router.GetRouter().AddAppRoutes(p.app.GetName(), routes)
}
