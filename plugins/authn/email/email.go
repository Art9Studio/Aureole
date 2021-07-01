package email

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	authzTypes "aureole/internal/plugins/authz/types"
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
	email struct {
		appName    string
		rawConf    *configs.Authn
		conf       *config
		identity   *identity.Identity
		storage    storageTypes.Storage
		coll       *collections.Collection
		authorizer authzTypes.Authorizer
		link       magicLink
	}

	magicLink struct {
		coll   *collections.Collection
		sender senderTypes.Sender
		hasher func() hash.Hash
	}
)

func (p *email) Init(appName string) (err error) {
	p.appName = appName

	p.conf, err = initConfig(&p.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi

	p.coll, err = pluginApi.Project.GetCollection(p.conf.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", p.conf.Collection)
	}

	p.link.coll, err = pluginApi.Project.GetCollection(p.conf.Link.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", p.conf.Link.Collection)
	}

	p.storage, err = pluginApi.Project.GetStorage(p.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", p.conf.Storage)
	}
	p.link.sender, err = pluginApi.Project.GetSender(p.conf.Link.Sender)
	if err != nil {
		return fmt.Errorf("sender named '%s' is not declared", p.conf.Link.Sender)
	}

	p.authorizer, err = pluginApi.Project.GetAuthorizer(p.rawConf.AuthzName, appName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", p.rawConf.AuthzName)
	}

	p.identity, err = pluginApi.Project.GetIdentity(appName)
	if err != nil {
		return fmt.Errorf("identity in app '%s' is not declared", appName)
	}

	p.link.hasher, err = initHasher(p.conf.Link.Token.HashFunc)
	if err != nil {
		return err
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
		return nil, fmt.Errorf("email auth: hasher '%s' doesn't supported", hasherName)
	}
	return h, nil
}

func createRoutes(p *email) {
	routes := []*_interface.Route{
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.Login.Path,
			Handler: GetMagicLink(p),
		},
		{
			Method:  "POST",
			Path:    p.rawConf.PathPrefix + p.conf.Register.Path,
			Handler: Register(p),
		},
		{
			Method:  "GET",
			Path:    p.rawConf.PathPrefix + p.conf.Link.Path,
			Handler: Login(p),
		},
	}
	authn.Repository.PluginApi.Router.AddAppRoutes(p.appName, routes)
}
