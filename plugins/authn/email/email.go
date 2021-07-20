package email

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	app "aureole/internal/context/interface"
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
		app        app.AppCtx
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

func (e *email) Init(app app.AppCtx) (err error) {
	e.app = app
	e.identity = app.GetIdentity()
	e.conf, err = initConfig(&e.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authn.Repository.PluginApi
	e.coll, err = pluginApi.Project.GetCollection(e.conf.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", e.conf.Collection)
	}

	e.link.coll, err = pluginApi.Project.GetCollection(e.conf.Link.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", e.conf.Link.Collection)
	}

	e.storage, err = pluginApi.Project.GetStorage(e.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", e.conf.Storage)
	}
	e.link.sender, err = pluginApi.Project.GetSender(e.conf.Link.Sender)
	if err != nil {
		return fmt.Errorf("sender named '%s' is not declared", e.conf.Link.Sender)
	}

	e.authorizer, err = e.app.GetAuthorizer(e.rawConf.AuthzName)
	if err != nil {
		return fmt.Errorf("authorizer named '%s' is not declared", e.rawConf.AuthzName)
	}

	e.link.hasher, err = initHasher(e.conf.Link.Token.HashFunc)
	if err != nil {
		return err
	}

	if err := e.storage.CheckFeaturesAvailable([]string{e.coll.Type}); err != nil {
		return err
	}

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

func createRoutes(e *email) {
	routes := []*_interface.Route{
		{
			Method:  "POST",
			Path:    e.rawConf.PathPrefix + e.conf.Login.Path,
			Handler: GetMagicLink(e),
		},
		{
			Method:  "POST",
			Path:    e.rawConf.PathPrefix + e.conf.Register.Path,
			Handler: Register(e),
		},
		{
			Method:  "GET",
			Path:    e.rawConf.PathPrefix + e.conf.Link.Path,
			Handler: Login(e),
		},
	}
	authn.Repository.PluginApi.Router.AddAppRoutes(e.app.GetName(), routes)
}
