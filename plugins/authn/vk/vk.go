package vk

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

const PluginID = "3888"

type vk struct {
	pluginApi  core.PluginAPI
	app        app.AppState
	rawConf    *configs.Authn
	conf       *config
	manager    identity.ManagerI
	provider   *oauth2.Config
	authorizer authzTypes.Authorizer
}

func (v *vk) Init(appName string, api core.PluginAPI) (err error) {
	v.pluginApi = api
	v.conf, err = initConfig(&v.rawConf.Config)
	if err != nil {
		return err
	}

	v.app, err = v.pluginApi.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app named '%s' is not declared", appName)
	}

	v.manager, err = v.app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared", appName)
	}

	v.authorizer, err = v.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", appName)
	}

	if err := initProvider(v); err != nil {
		return err
	}
	createRoutes(v)
	return nil
}

func (*vk) GetPluginID() string {
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

func initProvider(v *vk) error {
	redirectUri, err := v.app.GetUrl()
	if err != nil {
		return err
	}

	redirectUri.Path = path.Clean(redirectUri.Path + v.conf.PathPrefix + v.conf.RedirectUri)
	v.provider = &oauth2.Config{
		ClientID:     v.conf.ClientId,
		ClientSecret: v.conf.ClientSecret,
		Endpoint:     endpoints.Vk,
		RedirectURL:  redirectUri.String(),
		Scopes:       v.conf.Scopes,
	}
	return nil
}

func createRoutes(v *vk) {
	routes := []*_interface.Route{
		{
			Method:  "GET",
			Path:    v.conf.PathPrefix,
			Handler: GetAuthCode(v),
		},
		{
			Method:  "GET",
			Path:    v.conf.PathPrefix + v.conf.RedirectUri,
			Handler: Login(v),
		},
	}
	v.pluginApi.GetRouter().AddAppRoutes(v.app.GetName(), routes)
}
