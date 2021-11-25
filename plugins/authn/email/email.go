package email

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/core"
	senderTypes "aureole/internal/plugins/sender/types"
	"aureole/internal/router/interface"
	app "aureole/internal/state/interface"
	"fmt"
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

func (*email) GetPluginID() string {
	return PluginID
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
	routes := []*_interface.Route{
		{
			Method:  "POST",
			Path:    e.conf.PathPrefix + e.conf.SendUrl,
			Handler: SendMagicLink(e),
		},
		{
			Method:  "GET",
			Path:    e.conf.PathPrefix + e.conf.ConfirmUrl,
			Handler: Login(e),
		},
	}
	e.pluginApi.GetRouter().AddAppRoutes(e.app.GetName(), routes)
}
