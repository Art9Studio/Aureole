package context

import (
	"aureole/internal/collections"
	"aureole/internal/context/app"
	"aureole/internal/identity"
	authzTypes "aureole/internal/plugins/authz/types"
	cryptoKeyTypes "aureole/internal/plugins/cryptokey/types"
	pwhasherTypes "aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"fmt"
)

type (
	ProjectCtx struct {
		APIVersion  string
		Apps        map[string]*app.App
		Collections map[string]*collections.Collection
		Storages    map[string]storageTypes.Storage
		Hashers     map[string]pwhasherTypes.PwHasher
		Senders     map[string]senderTypes.Sender
		CryptoKeys  map[string]cryptoKeyTypes.CryptoKey
	}
)

func (ctx *ProjectCtx) GetCollection(name string) (*collections.Collection, error) {
	c, ok := ctx.Collections[name]
	if !ok {
		return nil, fmt.Errorf("can't find collection named '%s'", name)
	}

	return c, nil
}

func (ctx *ProjectCtx) GetStorage(name string) (storageTypes.Storage, error) {
	s, ok := ctx.Storages[name]
	if !ok {
		return nil, fmt.Errorf("can't find storage named '%s'", name)
	}

	return s, nil
}

func (ctx *ProjectCtx) GetHasher(name string) (pwhasherTypes.PwHasher, error) {
	h, ok := ctx.Hashers[name]
	if !ok {
		return nil, fmt.Errorf("can't find hasher named '%s'", name)
	}

	return h, nil
}

func (ctx *ProjectCtx) GetSender(name string) (senderTypes.Sender, error) {
	s, ok := ctx.Senders[name]
	if !ok {
		return nil, fmt.Errorf("can't find sender named '%s'", name)
	}

	return s, nil
}

func (ctx *ProjectCtx) GetCryptoKey(name string) (cryptoKeyTypes.CryptoKey, error) {
	k, ok := ctx.CryptoKeys[name]
	if !ok {
		return nil, fmt.Errorf("can't find crypto key named '%s'", name)
	}

	return k, nil
}

func (ctx *ProjectCtx) GetAuthorizer(name, appName string) (authzTypes.Authorizer, error) {
	app, ok := ctx.Apps[appName]
	if !ok {
		return nil, fmt.Errorf("can't find app named '%s'", appName)
	}

	authz, ok := app.Authorizers[name]
	if !ok {
		return nil, fmt.Errorf("can't find authorizer named '%s'", name)
	}

	return authz, nil
}

func (ctx *ProjectCtx) GetIdentity(appName string) (*identity.Identity, error) {
	app, ok := ctx.Apps[appName]
	if !ok {
		return nil, fmt.Errorf("can't find app named '%s'", appName)
	}

	return app.Identity, nil
}
