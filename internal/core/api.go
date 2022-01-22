package core

import (
	"aureole/internal/plugins"
)

type (
	PluginAPI struct {
		app       *App
		project   *Project
		router    *router
		keyPrefix string
	}

	option func(api *PluginAPI)
)

func initAPI(p *Project, options ...option) PluginAPI {
	api := PluginAPI{project: p}

	for _, option := range options {
		option(&api)
	}

	return api
}

func withKeyPrefix(prefix string) option {
	return func(api *PluginAPI) {
		api.keyPrefix = prefix
	}
}

func withApp(app *App) option {
	return func(api *PluginAPI) {
		api.app = app
	}
}

func withRouter(r *router) option {
	return func(api *PluginAPI) {
		api.router = r
	}
}

func (api PluginAPI) IsTestRun() bool {
	return api.project.IsTestRun()
}

func (api PluginAPI) Is2FAEnabled(cred *plugins.Credential, provider string) (bool, string, error) {
	manager, err := api.app.GetIDManager()
	if err != nil {
		return false, "", err
	}

	id, err := manager.GetData(cred, provider, plugins.SecondFactorID)
	if err != nil {
		return false, "", err
	}

	if id != "" {
		return true, id.(string), nil
	} else {
		return false, "", nil
	}
}

func (api PluginAPI) SaveToService(k string, v interface{}, exp int) error {
	serviceStorage, err := api.project.GetServiceStorage()
	if err != nil {
		return err
	}
	return serviceStorage.Set(api.keyPrefix+k, v, exp)
}

func (api PluginAPI) GetFromService(k string, v interface{}) (ok bool, err error) {
	serviceStorage, err := api.project.GetServiceStorage()
	if err != nil {
		return false, err
	}
	return serviceStorage.Get(api.keyPrefix+k, v)
}

func (api PluginAPI) GetApp(name string) (*App, error) {
	return api.project.GetApp(name)
}

func (api PluginAPI) GetAuthorizer(name string) (plugins.Authorizer, error) {
	return api.project.GetAuthorizer(name)
}

func (api PluginAPI) GetSecondFactor(name string) (plugins.SecondFactor, error) {
	return api.project.GetSecondFactor(name)
}

func (api PluginAPI) GetStorage(name string) (plugins.Storage, error) {
	return api.project.GetStorage(name)
}

func (api PluginAPI) GetKeyStorage(name string) (plugins.KeyStorage, error) {
	return api.project.GetKeyStorage(name)
}

func (api PluginAPI) GetHasher(name string) (plugins.PWHasher, error) {
	return api.project.GetHasher(name)
}

func (api PluginAPI) GetSender(name string) (plugins.Sender, error) {
	return api.project.GetSender(name)
}

func (api PluginAPI) GetCryptoKey(name string) (plugins.CryptoKey, error) {
	return api.project.GetCryptoKey(name)
}

func (api PluginAPI) AddAppRoutes(appName string, routes []*Route) {
	api.router.addAppRoutes(appName, routes)
}

func (api PluginAPI) AddProjectRoutes(routes []*Route) {
	api.router.addProjectRoutes(routes)
}

func (api PluginAPI) GetAppRoutes() map[string][]*Route {
	return api.router.getAppRoutes()
}

func (api PluginAPI) GetProjectRoutes() []*Route {
	return api.router.getProjectRoutes()
}
