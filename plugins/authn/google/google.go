package google

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

type google struct {
	app        app.AppState
	rawConf    *configs.Authn
	conf       *config
	manager    identity.ManagerI
	provider   *oauth2.Config
	authorizer authzTypes.Authorizer
}

func (g *google) Init(app app.AppState) (err error) {
	g.app = app
	g.rawConf.PathPrefix = "/oauth2/" + AdapterName

	g.conf, err = initConfig(&g.rawConf.Config)
	if err != nil {
		return err
	}

	g.manager, err = app.GetIdentityManager()
	if err != nil {
		fmt.Printf("manager for app '%s' is not declared, persist layer is not available", app.GetName())
	}

	g.authorizer, err = g.app.GetAuthorizer()
	if err != nil {
		return fmt.Errorf("authorizer named for app '%s' is not declared", app.GetName())
	}

	if err := initProvider(g); err != nil {
		return err
	}
	createRoutes(g)
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

func initProvider(g *google) error {
	redirectUri, err := g.app.GetUrl()
	if err != nil {
		return err
	}

	redirectUri.Path = path.Clean(redirectUri.Path + g.rawConf.PathPrefix + g.conf.RedirectUri)
	g.provider = &oauth2.Config{
		ClientID:     g.conf.ClientId,
		ClientSecret: g.conf.ClientSecret,
		Endpoint:     endpoints.Google,
		RedirectURL:  redirectUri.String(),
		Scopes:       g.conf.Scopes,
	}
	return nil
}

func createRoutes(g *google) {
	routes := []*_interface.Route{
		{
			Method:  "GET",
			Path:    g.rawConf.PathPrefix,
			Handler: GetAuthCode(g),
		},
		{
			Method:  "GET",
			Path:    g.rawConf.PathPrefix + g.conf.RedirectUri,
			Handler: Login(g),
		},
	}
	authn.Repository.PluginApi.Router.AddAppRoutes(g.app.GetName(), routes)
}
