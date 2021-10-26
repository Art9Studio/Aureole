package vk

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/router/interface"
	app "aureole/internal/state/interface"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"path"
)

type vk struct {
	app        app.AppState
	rawConf    *configs.Authn
	conf       *config
	manager    identity.ManagerI
	provider   *oauth2.Config
	authorizer authzTypes.Authorizer
}

func (v *vk) Init(app app.AppState) (err error) {
	v.app = app
	v.rawConf.PathPrefix = "/oauth2/" + AdapterName

	v.conf, err = initConfig(&v.rawConf.Config)
	if err != nil {
		return err
	}

	v.manager, err = app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared, persist layer is not available", app.GetName())
	}

	v.authorizer, err = v.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", app.GetName())
	}

	if err := initProvider(v); err != nil {
		return err
	}
	createRoutes(v)
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

func initProvider(v *vk) error {
	redirectUri, err := v.app.GetUrl()
	if err != nil {
		return err
	}

	redirectUri.Path = path.Clean(redirectUri.Path + v.rawConf.PathPrefix + v.conf.RedirectUri)
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
			Path:    v.rawConf.PathPrefix,
			Handler: GetAuthCode(v),
		},
		{
			Method:  "GET",
			Path:    v.rawConf.PathPrefix + v.conf.RedirectUri,
			Handler: Login(v),
		},
	}
	authn.Repository.PluginApi.Router.AddAppRoutes(v.app.GetName(), routes)
}
