package vk

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"aureole/internal/router/interface"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"net/url"
	"path"
)

const Provider = "vk"

type vk struct {
	appName    string
	appUrl     *url.URL
	rawConf    *configs.Authn
	conf       *config
	identity   *identity.Identity
	coll       *collections.Collection
	storage    storageT.Storage
	provider   *oauth2.Config
	authorizer authzTypes.Authorizer
}

func (v *vk) Init(appName string, appUrl *url.URL) (err error) {
	v.appName = appName
	v.appUrl = appUrl

	v.conf, err = initConfig(&v.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi
	v.identity, err = pluginApi.Project.GetIdentity(appName)
	if err != nil {
		return fmt.Errorf("identity in app '%s' is not declared", appName)
	}

	v.coll, err = pluginApi.Project.GetCollection(v.conf.Coll)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", v.conf.Coll)
	}

	v.storage, err = pluginApi.Project.GetStorage(v.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", v.conf.Storage)
	}

	v.authorizer, err = pluginApi.Project.GetAuthorizer(v.rawConf.AuthzName, appName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", v.rawConf.AuthzName)
	}

	redirectUri := v.appUrl
	redirectUri.Path = path.Clean(redirectUri.Path + v.rawConf.PathPrefix + v.conf.RedirectUri)
	v.provider = &oauth2.Config{
		ClientID:     v.conf.ClientId,
		ClientSecret: v.conf.ClientSecret,
		Endpoint:     endpoints.Vk,
		RedirectURL:  redirectUri.String(),
		Scopes:       v.conf.Scopes,
	}

	createRoutes(v)
	return nil
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}

	return adapterConf, nil
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
	authn.Repository.PluginApi.Router.AddAppRoutes(v.appName, routes)
}
