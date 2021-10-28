package email

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
	cKeyTypes "aureole/internal/plugins/cryptokey/types"
	senderTypes "aureole/internal/plugins/sender/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"aureole/internal/router/interface"
	app "aureole/internal/state/interface"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"net/url"
	"path"
)

type (
	email struct {
		app      app.AppState
		rawConf  *configs.Authn
		conf     *config
		identity *identity.Identity
		storage  storageTypes.Storage
		// coll       *collections.Collection
		authorizer authzTypes.Authorizer
		serviceKey cKeyTypes.CryptoKey
		sender     senderTypes.Sender
		magicLink  *url.URL
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
	e.identity, err = app.GetIdentity()
	if err != nil {
		return fmt.Errorf("identity for app '%s' is not declared", app.GetName())
	}

	e.serviceKey, err = pluginApi.Project.GetCryptoKey("service_internal_key")
	if err != nil {
		return errors.New("cryptokey named 'service_internal_key' is not declared")
	}

	/*e.coll, err = pluginApi.Project.GetCollection(e.conf.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", e.conf.Collection)
	}

	e.storage, err = pluginApi.Project.GetStorage(e.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", e.conf.Storage)
	}*/

	e.sender, err = pluginApi.Project.GetSender(e.conf.Sender)
	if err != nil {
		return fmt.Errorf("sender named '%s' is not declared", e.conf.Sender)
	}

	e.authorizer, err = e.app.GetAuthorizer(e.rawConf.AuthzName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", e.rawConf.AuthzName)
	}

	e.magicLink, err = createMagicLink(e)
	if err != nil {
		return err
	}

	/*if err := e.storage.CheckFeaturesAvailable([]string{e.coll.Type}); err != nil {
		return err
	}*/

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
	return u, nil
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
