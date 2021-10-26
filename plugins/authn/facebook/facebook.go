package facebook

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

type facebook struct {
	app        app.AppState
	rawConf    *configs.Authn
	conf       *config
	manager    identity.ManagerI
	provider   *oauth2.Config
	authorizer authzTypes.Authorizer
}

func (f *facebook) Init(app app.AppState) (err error) {
	f.app = app
	f.rawConf.PathPrefix = "/oauth2/" + AdapterName

	f.conf, err = initConfig(&f.rawConf.Config)
	if err != nil {
		return err
	}

	f.manager, err = app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared, persist layer is not available", app.GetName())
	}

	f.authorizer, err = f.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", app.GetName())
	}

	if err := initProvider(f); err != nil {
		return err
	}
	createRoutes(f)
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

func initProvider(f *facebook) error {
	redirectUri, err := f.app.GetUrl()
	if err != nil {
		return err
	}

	redirectUri.Path = path.Clean(redirectUri.Path + f.rawConf.PathPrefix + f.conf.RedirectUri)
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
			Path:    f.rawConf.PathPrefix,
			Handler: GetAuthCode(f),
		},
		{
			Method:  "GET",
			Path:    f.rawConf.PathPrefix + f.conf.RedirectUri,
			Handler: Login(f),
		},
	}
	authn.Repository.PluginApi.Router.AddAppRoutes(f.app.GetName(), routes)
}
