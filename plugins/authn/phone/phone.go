package phone

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"aureole/internal/router/interface"
	app "aureole/internal/state/interface"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type (
	phone struct {
		app          app.AppState
		rawConf      *configs.Authn
		conf         *config
		identity     *identity.Identity
		storage      storageTypes.Storage
		hasher       types.PwHasher
		coll         *collections.Collection
		authorizer   authzTypes.Authorizer
		verification verif
	}

	verif struct {
		coll   *collections.Collection
		sender senderTypes.Sender
	}
)

func (p *phone) Init(app app.AppState) (err error) {
	p.app = app
	p.rawConf.PathPrefix = "/" + AdapterName

	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi
	p.identity, err = app.GetIdentity()
	if err != nil {
		return fmt.Errorf("identity for app '%s' is not declared", app.GetName())
	}

	p.hasher, err = pluginApi.Project.GetHasher(p.conf.Hasher)
	if err != nil {
		return fmt.Errorf("hasher named '%s' is not declared", p.conf.Hasher)
	}

	/*p.coll, err = pluginApi.Project.GetCollection(p.conf.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", p.conf.Collection)
	}*/

	p.verification.coll, err = pluginApi.Project.GetCollection(p.conf.Verification.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", p.conf.Verification.Collection)
	}

	p.storage, err = pluginApi.Project.GetStorage(p.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", p.conf.Storage)
	}

	p.verification.sender, err = pluginApi.Project.GetSender(p.conf.Verification.Sender)
	if err != nil {
		return fmt.Errorf("sender named '%s' is not declared", p.conf.Verification.Sender)
	}

	p.authorizer, err = p.app.GetAuthorizer(p.rawConf.AuthzName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", p.rawConf.AuthzName)
	}

	if err := p.storage.CheckFeaturesAvailable([]string{p.coll.Type}); err != nil {
		return err
	}

	createRoutes(p)
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

func createRoutes(p *phone) {
	routes := []*_interface.Route{
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.Path,
			Handler: SendOtp(p),
		},
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.Verification.Path,
			Handler: Login(p),
		},
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.Verification.ResendUrl,
			Handler: Resend(p),
		},
	}
	authn.Repository.PluginApi.Router.AddAppRoutes(p.app.GetName(), routes)
}
