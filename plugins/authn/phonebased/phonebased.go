package phonebased

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
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type phoneBased struct {
	appName           string
	rawConf           *configs.Authn
	conf              *config
	identity          *identity.Identity
	storage           storageTypes.Storage
	hasher            types.PwHasher
	coll, confirmColl *collections.Collection
	authorizer        authzTypes.Authorizer
	sender            senderTypes.Sender
}

func (p *phoneBased) Init(appName string) (err error) {
	p.appName = appName

	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi
	p.hasher, err = pluginApi.Project.GetHasher(p.conf.Hasher)
	if err != nil {
		return fmt.Errorf("hasher named '%s' is not declared", p.conf.Hasher)
	}

	p.coll, err = pluginApi.Project.GetCollection(p.conf.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", p.conf.Collection)
	}

	p.confirmColl, err = pluginApi.Project.GetCollection(p.conf.VerificationColl)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", p.conf.VerificationColl)
	}

	p.storage, err = pluginApi.Project.GetStorage(p.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", p.conf.Storage)
	}
	p.sender, err = pluginApi.Project.GetSender(p.conf.Sender)
	if err != nil {
		return fmt.Errorf("sender named '%s' is not declared", p.conf.Sender)
	}

	p.authorizer, err = pluginApi.Project.GetAuthorizer(p.rawConf.AuthzName, appName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", p.rawConf.AuthzName)
	}

	p.identity, err = pluginApi.Project.GetIdentity(appName)
	if err != nil {
		return fmt.Errorf("identity in app '%s' is not declared", appName)
	}

	if err = p.storage.CheckFeaturesAvailable([]string{p.coll.Type}); err != nil {
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

func createRoutes(p *phoneBased) {
	routes := []*_interface.Route{
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.Login.Path,
			Handler: Login(p),
		},
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.Register.Path,
			Handler: Register(p),
		},
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.Verification.Path,
			Handler: Confirm(p),
		},
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.ResendUrl,
			Handler: Resend(p),
		},
	}
	authn.Repository.PluginApi.Router.AddAppRoutes(p.appName, routes)
}
