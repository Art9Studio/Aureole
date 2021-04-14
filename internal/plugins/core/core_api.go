package core

import (
	contextTypes "aureole/context/types"
	"aureole/internal/collections"
	authzTypes "aureole/internal/plugins/authz/types"
	ckeyTypes "aureole/internal/plugins/cryptokey/types"
	pwhasherTypes "aureole/internal/plugins/pwhasher/types"
	senderTypes "aureole/internal/plugins/sender/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"fmt"
)

type PluginApi struct {
	projectCtx *contextTypes.ProjectCtx
}

func InitPluginApi(ctx *contextTypes.ProjectCtx) *PluginApi {
	return &PluginApi{projectCtx: ctx}
}

func (api *PluginApi) GetCollection(name string) (*collections.Collection, error) {
	c, ok := api.projectCtx.Collections[name]
	if !ok {
		return nil, fmt.Errorf("can't find collection named '%s'", name)
	}

	return c, nil
}

func (api *PluginApi) GetStorage(name string) (storageTypes.Storage, error) {
	s, ok := api.projectCtx.Storages[name]
	if !ok {
		return nil, fmt.Errorf("can't find storage named '%s'", name)
	}

	return s, nil
}

func (api *PluginApi) GetHasher(name string) (pwhasherTypes.PwHasher, error) {
	h, ok := api.projectCtx.Hashers[name]
	if !ok {
		return nil, fmt.Errorf("can't find hasher named '%s'", name)
	}

	return h, nil
}

func (api *PluginApi) GetSender(name string) (senderTypes.Sender, error) {
	s, ok := api.projectCtx.Senders[name]
	if !ok {
		return nil, fmt.Errorf("can't find sender named '%s'", name)
	}

	return s, nil
}

func (api *PluginApi) GetCryptoKey(name string) (ckeyTypes.CryptoKey, error) {
	k, ok := api.projectCtx.CryptoKeys[name]
	if !ok {
		return nil, fmt.Errorf("can't find crypto key named '%s'", name)
	}

	return k, nil
}

func (api *PluginApi) GetAuthorizer(name, appName string) (authzTypes.Authorizer, error) {
	app, ok := api.projectCtx.Apps[appName]
	if !ok {
		return nil, fmt.Errorf("can't find app named '%s'", appName)
	}

	authz, ok := app.Authorizers[name]
	if !ok {
		return nil, fmt.Errorf("can't find authorizer named '%s'", name)
	}

	return authz, nil
}
