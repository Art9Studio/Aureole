package facebook

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/core"
	"aureole/internal/router/interface"
	app "aureole/internal/state/interface"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"path"
)

const PluginID = "3030"

type facebook struct {
	pluginApi  core.PluginAPI
	app        app.AppState
	rawConf    *configs.Authn
	conf       *config
	manager    identity.ManagerI
	provider   *oauth2.Config
	authorizer authzTypes.Authorizer
}

func (f *facebook) Init(appName string, api core.PluginAPI) (err error) {
	f.pluginApi = api
	f.conf, err = initConfig(&f.rawConf.Config)
	if err != nil {
		return err
	}

	f.app, err = f.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	f.manager, err = f.app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared, persist layer is not available", appName)
	}

	f.authorizer, err = f.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", appName)
	}

	if err := initProvider(f); err != nil {
		return err
	}
	createRoutes(f)
	return nil
}

func (*facebook) GetPluginID() string {
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

func initProvider(f *facebook) error {
	redirectUri, err := f.app.GetUrl()
	if err != nil {
		return err
	}

	redirectUri.Path = path.Clean(redirectUri.Path + f.conf.PathPrefix + f.conf.RedirectUri)
	f.provider = &oauth2.Config{
		ClientID:     f.conf.ClientId,
		ClientSecret: f.conf.ClientSecret,
		Endpoint:     endpoints.Facebook,
		RedirectURL:  redirectUri.String(),
		Scopes:       f.conf.Scopes,
	}
	return nil
}

func createRoutes(f *facebook) {
	routes := []*_interface.Route{
		{
			Method:  "GET",
			Path:    f.conf.PathPrefix,
			Handler: GetAuthCode(f),
		},
		{
			Method:  "GET",
			Path:    f.conf.PathPrefix + f.conf.RedirectUri,
			Handler: Login(f),
		},
	}
	f.pluginApi.GetRouter().AddAppRoutes(f.app.GetName(), routes)
}
