package core

import (
	"aureole/internal/identity"
	"aureole/internal/plugins/2fa/types"
	authzT "aureole/internal/plugins/authz/types"
	cryptoKeyT "aureole/internal/plugins/cryptokey/types"
	kstorageT "aureole/internal/plugins/kstorage/types"
	pwhasherT "aureole/internal/plugins/pwhasher/types"
	senderT "aureole/internal/plugins/sender/types"
	storageT "aureole/internal/plugins/storage/types"
	state "aureole/internal/state/interface"
)

type (
	PluginAPI interface {
		IsTestRun() bool
		Is2FAEnabled(credential *identity.Credential, provider string) (ok bool, id string, err error)
		SaveToService(k string, v interface{}, exp int) error
		GetFromService(k string, v interface{}) (ok bool, err error)
		GetApp(name string) (state.AppState, error)
		GetAuthorizer(name string) (authzT.Authorizer, error)
		GetSecondFactor(name string) (types.SecondFactor, error)
		GetStorage(name string) (storageT.Storage, error)
		GetKeyStorage(name string) (kstorageT.KeyStorage, error)
		GetHasher(name string) (pwhasherT.PwHasher, error)
		GetSender(name string) (senderT.Sender, error)
		GetCryptoKey(name string) (cryptoKeyT.CryptoKey, error)
	}

	pluginAPI struct {
		app       state.AppState
		project   state.ProjectState
		keyPrefix string
	}

	APIOption func(api *pluginAPI)
)

func InitAPI(p state.ProjectState, options ...APIOption) PluginAPI {
	api := pluginAPI{project: p}

	for _, option := range options {
		option(&api)
	}

	return api
}

func WithKeyPrefix(prefix string) APIOption {
	return func(api *pluginAPI) {
		api.keyPrefix = prefix
	}
}

func WithAppState(app state.AppState) APIOption {
	return func(api *pluginAPI) {
		api.app = app
	}
}

func (api pluginAPI) IsTestRun() bool {
	return api.project.IsTestRun()
}

func (api pluginAPI) Is2FAEnabled(cred *identity.Credential, provider string) (bool, string, error) {
	manager, err := api.app.GetIdentityManager()
	if err != nil {
		return false, "", err
	}

	id, err := manager.GetData(cred, provider, identity.SecondFactorID)
	if err != nil {
		return false, "", err
	}

	if id != "" {
		return true, id.(string), nil
	} else {
		return false, "", nil
	}
}

func (api pluginAPI) SaveToService(k string, v interface{}, exp int) error {
	serviceStorage, err := api.project.GetServiceStorage()
	if err != nil {
		return err
	}
	return serviceStorage.Set(api.keyPrefix+k, v, exp)
}

func (api pluginAPI) GetFromService(k string, v interface{}) (ok bool, err error) {
	serviceStorage, err := api.project.GetServiceStorage()
	if err != nil {
		return false, err
	}
	return serviceStorage.Get(api.keyPrefix+k, v)
}

func (api pluginAPI) GetApp(name string) (state.AppState, error) {
	return api.project.GetApp(name)
}

func (api pluginAPI) GetAuthorizer(name string) (authzT.Authorizer, error) {
	return api.project.GetAuthorizer(name)
}

func (api pluginAPI) GetSecondFactor(name string) (types.SecondFactor, error) {
	return api.project.GetSecondFactor(name)
}

func (api pluginAPI) GetStorage(name string) (storageT.Storage, error) {
	return api.project.GetStorage(name)
}

func (api pluginAPI) GetKeyStorage(name string) (kstorageT.KeyStorage, error) {
	return api.project.GetKeyStorage(name)
}

func (api pluginAPI) GetHasher(name string) (pwhasherT.PwHasher, error) {
	return api.project.GetHasher(name)
}

func (api pluginAPI) GetSender(name string) (senderT.Sender, error) {
	return api.project.GetSender(name)
}

func (api pluginAPI) GetCryptoKey(name string) (cryptoKeyT.CryptoKey, error) {
	return api.project.GetCryptoKey(name)
}
