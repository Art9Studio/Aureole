package google

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"aureole/internal/router/interface"
	app "aureole/internal/state/interface"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"path"
)

const Provider = "google"

type google struct {
	app      app.AppState
	rawConf  *configs.Authn
	conf     *config
	identity *identity.Identity
	// coll       *collections.Collection
	storage    storageT.Storage
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

	// pluginApi := authn.Repository.PluginApi
	g.identity, err = app.GetIdentity()
	if err != nil {
		return fmt.Errorf("identity for app '%s' is not declared", app.GetName())
	}

	/*g.coll, err = pluginApi.Project.GetCollection(g.conf.Coll)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", g.conf.Coll)
	}

	g.storage, err = pluginApi.Project.GetStorage(g.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", g.conf.Storage)
	}*/

	g.authorizer, err = g.app.GetAuthorizer(g.rawConf.AuthzName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", g.rawConf.AuthzName)
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
