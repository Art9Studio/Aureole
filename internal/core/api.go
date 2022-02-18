package core

import (
	"aureole/internal/plugins"
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/jwt"
	"net/url"
	"regexp"
)

type (
	PluginAPI struct {
		app       *app
		project   *project
		router    *router
		keyPrefix string
	}

	option func(api *PluginAPI)
)

func initAPI(options ...option) PluginAPI {
	api := PluginAPI{}

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

func withProject(project *project) option {
	return func(api *PluginAPI) {
		api.project = project
	}
}

func withApp(app *app) option {
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
	return api.project.testRun
}

func (api PluginAPI) Is2FAEnabled(cred *plugins.Credential, mfaID string) (bool, error) {
	manager, ok := api.app.getIDManager()
	if !ok {
		return false, fmt.Errorf("can't find identity manager for app %s", api.app.name)
	}

	mfaData, err := manager.Get2FAData(cred, mfaID)
	if err != nil && !errors.Is(err, plugins.UserNotExistError) {
		return false, err
	}
	if mfaData != nil {
		return true, nil
	} else {
		return false, nil
	}
}

func (api PluginAPI) GetAppName() string {
	return api.app.name
}

func (api PluginAPI) GetAppUrl() url.URL {
	return *api.app.url
}

func (api PluginAPI) GetAppPathPrefix() string {
	return api.app.pathPrefix
}

func (api PluginAPI) GetAuthSessionExp() int {
	return api.app.authSessionExp
}

func (api PluginAPI) GetAuthorizer() (plugins.Authorizer, bool) {
	return api.app.getAuthorizer()
}

func (api PluginAPI) GetSecondFactors() (map[string]plugins.SecondFactor, bool) {
	return api.app.getSecondFactors()
}

func (api PluginAPI) GetStorage(name string) (plugins.Storage, bool) {
	return api.app.getStorage(name)
}

func (api PluginAPI) GetIDManager() (plugins.IDManager, bool) {
	return api.app.getIDManager()
}

func (api PluginAPI) GetCryptoStorage(name string) (plugins.CryptoStorage, bool) {
	return api.app.getCryptoStorage(name)
}

func (api PluginAPI) GetSender(name string) (plugins.Sender, bool) {
	return api.app.getSender(name)
}

func (api PluginAPI) GetCryptoKey(name string) (plugins.CryptoKey, bool) {
	return api.app.getCryptoKey(name)
}

func (api PluginAPI) AddAppRoutes(routes []*Route) {
	api.router.addAppRoutes(api.app.name, routes)
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

func (api PluginAPI) SaveToService(k string, v interface{}, exp int) error {
	serviceStorage, ok := api.app.getServiceStorage()
	if !ok {
		return errors.New("can't find service storage")
	}
	return serviceStorage.Set(api.keyPrefix+k, v, exp)
}

func (api PluginAPI) GetFromService(k string, v interface{}) (ok bool, err error) {
	serviceStorage, ok := api.app.getServiceStorage()
	if !ok {
		return false, errors.New("can't find service storage")
	}
	return serviceStorage.Get(api.keyPrefix+k, v)
}

func (api PluginAPI) Encrypt(data interface{}) ([]byte, error) {
	return encrypt(api.app, data)
}

func (api PluginAPI) Decrypt(data []byte, value interface{}) error {
	return decrypt(api.app, data, value)
}

func (api PluginAPI) GetRandStr(length int, alphabet string) (string, error) {
	return getRandStr(length, alphabet)
}

func (api PluginAPI) CreateJWT(payload map[string]interface{}, exp int) (string, error) {
	return createJWT(api.app, payload, exp)
}

func (api PluginAPI) ParseJWT(rawToken string) (jwt.Token, error) {
	return parseJWT(api.app, rawToken)
}

func (api PluginAPI) InvalidateJWT(token jwt.Token) error {
	return invalidateJWT(api.app, token)
}

func (api PluginAPI) GetFromJWT(token jwt.Token, name string, value interface{}) error {
	return getFromJWT(token, name, value)
}

func (api PluginAPI) Filter(fields, filters map[string]string) (bool, error) {
	for fieldName, pattern := range filters {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false, err
		}
		if !re.MatchString(fields[fieldName]) {
			return false, nil
		}
	}
	return true, nil
}
