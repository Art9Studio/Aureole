package email

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
	senderTypes "aureole/internal/plugins/sender/types"
	"aureole/internal/router/interface"
	app "aureole/internal/state/interface"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"net/url"
	"path"
)

type (
	email struct {
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

func (e *email) Init(app app.AppState) (err error) {
	e.app = app
	e.rawConf.PathPrefix = "/email-link"

	e.conf, err = initConfig(&e.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi
	e.manager, err = app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared, the persist layer is not available", app.GetName())
	}

	e.sender, err = pluginApi.Project.GetSender(e.conf.Sender)
	if err != nil {
		return fmt.Errorf("sender named '%s' is not declared", e.conf.Sender)
	}

	e.authorizer, err = e.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", app.GetName())
	}

	e.magicLink, err = createMagicLink(e)
	if err != nil {
		return err
	}

	createRoutes(e)
	return nil
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

	u.Path = path.Clean(u.Path + e.rawConf.PathPrefix + e.conf.ConfirmUrl)
	return &u, nil
}

func createRoutes(e *email) {
	routes := []*_interface.Route{
		{
			Method:  "POST",
			Path:    e.rawConf.PathPrefix + e.conf.SendUrl,
			Handler: SendMagicLink(e),
		},
		{
			Method:  "GET",
			Path:    e.rawConf.PathPrefix + e.conf.ConfirmUrl,
			Handler: Login(e),
		},
	}
	authn.Repository.PluginApi.Router.AddAppRoutes(e.app.GetName(), routes)
}
