package email

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/jwt"
	"aureole/internal/plugins"
	authnT "aureole/internal/plugins/authn/types"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/core"
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

const PluginID = "4071"

type (
	email struct {
		pluginApi  core.PluginAPI
		app        app.AppState
		rawConf    *configs.Authn
		conf       *config
		manager    identity.ManagerI
		authorizer authzTypes.Authorizer
		sender     senderTypes.Sender
		magicLink  *url.URL
	}

	input struct {
		Email string `json:"email"`
	}
)

func (e *email) Init(appName string, api core.PluginAPI) (err error) {
	e.pluginApi = api
	e.conf, err = initConfig(&e.rawConf.Config)
	if err != nil {
		return err
	}

	e.app, err = e.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	e.manager, err = e.app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared", appName)
	}

	e.sender, err = e.pluginApi.GetSender(e.conf.Sender)
	if err != nil {
		return fmt.Errorf("sender named '%s' is not declared", e.conf.Sender)
	}

	e.authorizer, err = e.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", appName)
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
		Type: AdapterName,
		ID:   PluginID,
	}
}

func (*email) Login() authnT.AuthFunc {
	return func(c fiber.Ctx) (*identity.Credential, fiber.Map, error) {
		rawToken := c.Query("token")
		if rawToken == "" {
			return nil, nil, errors.New("token not found")
		}

		token, err := jwt.ParseJWT(rawToken)
		if err != nil {
			return nil, nil, errors.New(err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return nil, nil, errors.New("cannot get email from token")
		}
		if err := jwt.InvalidateJWT(token); err != nil {
			return nil, nil, errors.New(err.Error())
		}

		return &identity.Credential{
				Name:  identity.Email,
				Value: email.(string),
			},
			fiber.Map{identity.Email: email}, nil
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
	u, err := e.app.GetUrl()
	if err != nil {
		return nil, err
	}

	u.Path = path.Clean(u.Path + e.conf.PathPrefix + e.conf.ConfirmUrl)
	return &u, nil
}

func createRoutes(e *email) {
	routes := []*router.Route{
		{
			Method:  router.MethodPOST,
			Path:    e.conf.PathPrefix + e.conf.SendUrl,
			Handler: SendMagicLink(e),
		},
	}
	router.GetRouter().AddAppRoutes(e.app.GetName(), routes)
}
