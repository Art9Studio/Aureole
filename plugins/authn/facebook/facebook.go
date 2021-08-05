package facebook

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	app "aureole/internal/context/interface"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"aureole/internal/router/interface"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"path"
)

const Provider = "facebook"

type facebook struct {
	app        app.AppCtx
	rawConf    *configs.Authn
	conf       *config
	identity   *identity.Identity
	coll       *collections.Collection
	storage    storageT.Storage
	provider   *oauth2.Config
	authorizer authzTypes.Authorizer
}

func (f *facebook) Init(app app.AppCtx) (err error) {
	f.app = app
	f.identity = app.GetIdentity()
	f.conf, err = initConfig(&f.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi
	f.coll, err = pluginApi.Project.GetCollection(f.conf.Coll)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", f.conf.Coll)
	}

	f.storage, err = pluginApi.Project.GetStorage(f.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", f.conf.Storage)
	}

	f.authorizer, err = f.app.GetAuthorizer(f.rawConf.AuthzName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", f.rawConf.AuthzName)
	}

	initProvider(f)
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

func initProvider(f *facebook) {
	redirectUri := f.app.GetUrl()
	redirectUri.Path = path.Clean(redirectUri.Path + f.rawConf.PathPrefix + f.conf.RedirectUri)
	f.provider = &oauth2.Config{
		ClientID:     f.conf.ClientId,
		ClientSecret: f.conf.ClientSecret,
		Endpoint:     endpoints.Facebook,
		RedirectURL:  redirectUri.String(),
		Scopes:       f.conf.Scopes,
	}
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
