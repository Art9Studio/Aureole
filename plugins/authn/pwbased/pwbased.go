package pwbased

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
	cKeyTypes "aureole/internal/plugins/cryptokey/types"
	"aureole/internal/plugins/pwhasher/types"
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
	pwBased struct {
		app        app.AppState
		rawConf    *configs.Authn
		conf       *config
		identity   *identity.Identity
		pwHasher   types.PwHasher
		storage    storageTypes.Storage
		coll       *collections.Collection
		authorizer authzTypes.Authorizer
		serviceKey cKeyTypes.CryptoKey
		reset      *reset
		verif      *verification
	}

	reset struct {
		sender      senderTypes.Sender
		confirmLink *url.URL
	}

	verification struct {
		sender      senderTypes.Sender
		confirmLink *url.URL
	}

	linkType string
)

const (
	ResetLink  linkType = "reset"
	VerifyLink linkType = "verify"
)

func (p *pwBased) Init(app app.AppState) (err error) {
	p.app = app
	p.rawConf.PathPrefix = "/"

	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi
	p.identity, err = app.GetIdentity()
	if err != nil {
		return fmt.Errorf("identity for app '%s' is not declared", app.GetName())
	}

	p.pwHasher, err = pluginApi.Project.GetHasher(p.conf.MainHasher)
	if err != nil {
		return fmt.Errorf("hasher named '%s' is not declared", p.conf.MainHasher)
	}

	p.serviceKey, err = pluginApi.Project.GetCryptoKey("service_internal_key")
	if err != nil {
		return errors.New("cryptokey named 'service_internal_key' is not declared")
	}

	/*p.coll, err = pluginApi.Project.GetCollection(p.conf.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", p.conf.Collection)
	}*/

	p.storage, err = pluginApi.Project.GetStorage(p.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", p.conf.Storage)
	}

	p.authorizer, err = p.app.GetAuthorizer(p.rawConf.AuthzName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", p.rawConf.AuthzName)
	}

	if pwResetEnable(p) {
		p.reset = &reset{}
		p.reset.sender, err = pluginApi.Project.GetSender(p.conf.Reset.Sender)
		if err != nil {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Reset.Sender)
		}

		p.reset.confirmLink, err = createConfirmLink(ResetLink, p)
		if err != nil {
			return err
		}
	}

	if verifEnable(p) {
		p.verif = &verification{}
		p.verif.sender, err = pluginApi.Project.GetSender(p.conf.Verif.Sender)
		if err != nil {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Verif.Sender)
		}

		p.verif.confirmLink, err = createConfirmLink(VerifyLink, p)
		if err != nil {
			return err
		}
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

func pwResetEnable(p *pwBased) bool {
	return p.conf.Reset.Sender != "" && p.conf.Reset.Template != ""
}

func verifEnable(p *pwBased) bool {
	return p.conf.Verif.Sender != "" && p.conf.Verif.Template != ""
}

func createConfirmLink(linkType linkType, p *pwBased) (*url.URL, error) {
	u, err := p.app.GetUrl()
	if err != nil {
		return nil, err
	}

	switch linkType {
	case ResetLink:
		u.Path = path.Clean(u.Path + p.rawConf.PathPrefix + p.conf.Reset.ConfirmUrl)
	case VerifyLink:
		u.Path = path.Clean(u.Path + p.rawConf.PathPrefix + p.conf.Verif.ConfirmUrl)
	}

	return u, nil
}

func createRoutes(p *pwBased) {
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
	}

	if pwResetEnable(p) {
		resetRoutes := []*_interface.Route{
			{
				Method:  "POST",
				Path:    p.rawConf.PathPrefix + p.conf.Reset.Path,
				Handler: Reset(p),
			},
			{
				Method:  "POST",
				Path:    p.rawConf.PathPrefix + p.conf.Reset.ConfirmUrl,
				Handler: ResetConfirm(p),
			},
		}
		routes = append(routes, resetRoutes...)
	}

	if verifEnable(p) {
		verifRoutes := []*_interface.Route{
			{
				Method:  "POST",
				Path:    p.rawConf.PathPrefix + p.conf.Verif.Path,
				Handler: Verify(p),
			},
			{
				Method:  "GET",
				Path:    p.rawConf.PathPrefix + p.conf.Verif.ConfirmUrl,
				Handler: VerifyConfirm(p),
			},
		}
		routes = append(routes, verifRoutes...)
	}

	authn.Repository.PluginApi.Router.AddAppRoutes(p.app.GetName(), routes)
}
