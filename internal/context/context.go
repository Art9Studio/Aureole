package context

import (
	"aureole/internal/collections"
	"aureole/internal/context/app"
	adminTypes "aureole/internal/plugins/admin/types"
	cryptoKeyTypes "aureole/internal/plugins/cryptokey/types"
	pwhasherTypes "aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"fmt"
)

type (
	ProjectCtx struct {
		APIVersion  string
		TestRun     bool
		PingPath    string
		Apps        map[string]*app.App
		Collections map[string]*collections.Collection
		Storages    map[string]storageTypes.Storage
		Hashers     map[string]pwhasherTypes.PwHasher
		Senders     map[string]senderTypes.Sender
		CryptoKeys  map[string]cryptoKeyTypes.CryptoKey
		Admins      map[string]adminTypes.Admin
	}
)

func (ctx *ProjectCtx) IsTestRun() bool {
	return ctx.TestRun
}

func (ctx *ProjectCtx) GetApp(name string) (*app.App, error) {
	a, ok := ctx.Apps[name]
	if !ok {
		return nil, fmt.Errorf("can't find app named '%s'", name)
	}

	return a, nil
}

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
