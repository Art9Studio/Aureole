package pwbased

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	app "aureole/internal/context/interface"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"aureole/internal/router/interface"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"hash"
)

type (
	pwBased struct {
		app        app.AppCtx
		rawConf    *configs.Authn
		conf       *config
		identity   *identity.Identity
		pwHasher   types.PwHasher
		storage    storageTypes.Storage
		coll       *collections.Collection
		authorizer authzTypes.Authorizer
		reset      *reset
		verif      *verification
	}

	reset struct {
		coll   *collections.Collection
		sender senderTypes.Sender
		hasher func() hash.Hash
	}

	verification struct {
		coll   *collections.Collection
		sender senderTypes.Sender
		hasher func() hash.Hash
	}

	linkType string
)

const (
	ResetLink  linkType = "reset"
	VerifyLink linkType = "verify"
)

func (p *pwBased) Init(app app.AppCtx) (err error) {
	p.app = app
	p.identity = app.GetIdentity()
	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi
	p.pwHasher, err = pluginApi.Project.GetHasher(p.conf.MainHasher)
	if err != nil {
		return fmt.Errorf("hasher named '%s' is not declared", p.conf.MainHasher)
	}

	p.coll, err = pluginApi.Project.GetCollection(p.conf.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", p.conf.Collection)
	}

	p.storage, err = pluginApi.Project.GetStorage(p.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", p.conf.Storage)
	}

	p.authorizer, err = p.app.GetAuthorizer(p.rawConf.AuthzName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", p.rawConf.AuthzName)
	}

	storageFeatures := []string{p.coll.Type}

	if pwResetEnable(p) {
		p.reset = &reset{}
		p.reset.coll, err = pluginApi.Project.GetCollection(p.conf.Reset.Collection)
		if err != nil {
			return fmt.Errorf("collection named '%s' is not declared", p.conf.Reset.Collection)
		}

		p.reset.sender, err = pluginApi.Project.GetSender(p.conf.Reset.Sender)
		if err != nil {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Reset.Sender)
		}

		p.reset.hasher, err = initHasher(p.conf.Reset.Token.HashFunc)
		if err != nil {
			return err
		}

		storageFeatures = append(storageFeatures, p.reset.coll.Type)
	}

	if verifEnable(p) {
		p.verif = &verification{}
		p.verif.coll, err = pluginApi.Project.GetCollection(p.conf.Verif.Collection)
		if err != nil {
			return fmt.Errorf("collection named '%s' is not declared", p.conf.Verif.Collection)
		}

		p.verif.sender, err = pluginApi.Project.GetSender(p.conf.Verif.Sender)
		if err != nil {
			return fmt.Errorf("sender named '%s' is not declared", p.conf.Verif.Sender)
		}

		p.verif.hasher, err = initHasher(p.conf.Verif.Token.HashFunc)
		if err != nil {
			return err
		}

		storageFeatures = append(storageFeatures, p.verif.coll.Type)
	}

	if err := p.storage.CheckFeaturesAvailable(storageFeatures); err != nil {
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
	return p.conf.Reset.Collection != "" && p.conf.Reset.Sender != "" && p.conf.Reset.Template != ""
}

func verifEnable(p *pwBased) bool {
	return p.conf.Verif.Collection != "" && p.conf.Verif.Sender != "" && p.conf.Verif.Template != ""
}

func initHasher(hasherName string) (func() hash.Hash, error) {
	var h func() hash.Hash
	switch hasherName {
	case "sha1":
		h = sha1.New
	case "sha224":
		h = sha256.New224
	case "sha256":
		h = sha256.New
	case "sha384":
		h = sha512.New384
	case "sha512":
		h = sha512.New
	default:
		return nil, fmt.Errorf("hasher '%s' doesn't supported", hasherName)
	}
	return h, nil
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
