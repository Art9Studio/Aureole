package core

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins"
	"crypto/tls"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"unicode/utf8"
)

var p *Project

type (
	PluginInitializer interface {
		Init(api PluginAPI) error
	}

	AppPluginInitializer interface {
		Init(appName string, api PluginAPI) error
	}
)

func Init(conf *configs.Project) {
	p = &Project{
		apiVersion: conf.APIVersion,
		testRun:    conf.TestRun,
		pingPath:   conf.PingPath,
	}

	createGlobalPlugins(conf, p)
	createApps(conf, p)
	createAppPlugins(conf, p)

	initGlobalPlugins(p)
	initAppPlugins(p)

	var err error
	if p.service.signKey, err = p.GetCryptoKey(conf.Service.SignKey); err != nil {
		fmt.Printf("cannot init service key: %v\n", err)
	}
	if p.service.encKey, err = p.GetCryptoKey(conf.Service.EncKey); err != nil {
		fmt.Printf("cannot init service key: %v\n", err)
	}
	if p.service.storage, err = p.GetStorage(conf.Service.Storage); err != nil {
		fmt.Printf("cannot init service storage: %v\n", err)
	}

	if p.testRun {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	getRouter().addProjectRoutes([]*Route{
		{
			Method: "GET",
			Path:   p.pingPath,
			Handler: func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			},
		},
	})

	listPluginStatus()
}

func createGlobalPlugins(conf *configs.Project, p *Project) {
	createAuthorizers(conf, p)
	createSecondFactors(conf, p)
	createPwHashers(conf, p)
	createSenders(conf, p)
	createCryptoKeys(conf, p)
	createStorages(conf, p)
	createKeyStorages(conf, p)
	createAdmins(conf, p)
}

func createAuthorizers(conf *configs.Project, p *Project) {
	p.authorizers = make(map[string]plugins.Authorizer)

	for i := range conf.Authz {
		authzConf := conf.Authz[i]
		a, err := plugins.NewAuthZ(&authzConf)
		if err != nil {
			fmt.Printf("cannot create authorizator '%s': %v\n", authzConf.Name, err)
		}

		p.authorizers[authzConf.Name] = a
	}
}

func createSecondFactors(conf *configs.Project, p *Project) {
	p.secondFactors = make(map[string]plugins.SecondFactor)

	for i := range conf.SecondFactors {
		mfaConf := conf.SecondFactors[i]
		secondFactor, err := plugins.NewSecondFactor(&mfaConf)
		if err != nil {
			fmt.Printf("cannot create second factor '%s': %v\n", mfaConf.Name, err)
		}

		p.secondFactors[mfaConf.Name] = secondFactor
	}
}

func createPwHashers(conf *configs.Project, p *Project) {
	p.hashers = make(map[string]plugins.PWHasher)

	for i := range conf.HasherConfs {
		hasherConf := conf.HasherConfs[i]
		h, err := plugins.NewPWHasher(&conf.HasherConfs[i])
		if err != nil {
			fmt.Printf("cannot create hasher '%s': %v\n", hasherConf.Name, err)
		}

		p.hashers[hasherConf.Name] = h
	}
}

func createSenders(conf *configs.Project, p *Project) {
	p.senders = make(map[string]plugins.Sender)

	for i := range conf.Senders {
		senderConf := conf.Senders[i]
		s, err := plugins.NewSender(&senderConf)
		if err != nil {
			fmt.Printf("cannot create sender '%s': %v\n", senderConf.Name, err)
		}

		p.senders[senderConf.Name] = s
	}
}

func createCryptoKeys(conf *configs.Project, p *Project) {
	p.cryptoKeys = make(map[string]plugins.CryptoKey)

	for i := range conf.CryptoKeys {
		ckeyConf := conf.CryptoKeys[i]
		ckey, err := plugins.NewCryptoKey(&ckeyConf)
		if err != nil {
			fmt.Printf("cannot create crypto key '%s': %v\n", ckeyConf.Name, err)
		}

		p.cryptoKeys[ckeyConf.Name] = ckey
	}
}

func createStorages(conf *configs.Project, p *Project) {
	p.storages = make(map[string]plugins.Storage)

	for i := range conf.Storages {
		storageConf := conf.Storages[i]
		s, err := plugins.NewStorage(&storageConf)
		if err != nil {
			fmt.Printf("open connection session to storage '%s': %v\n", storageConf.Name, err)
		}

		p.storages[storageConf.Name] = s
	}
}

func createKeyStorages(conf *configs.Project, p *Project) {
	p.keyStorages = make(map[string]plugins.KeyStorage)

	for i := range conf.KeyStorages {
		storageConf := conf.KeyStorages[i]
		s, err := plugins.NewKeyStorage(&storageConf)
		if err != nil {
			fmt.Printf("open connection session to key storage '%s': %v\n", storageConf.Name, err)
		}

		p.keyStorages[storageConf.Name] = s
	}
}

func createAdmins(conf *configs.Project, p *Project) {
	p.admins = make(map[string]plugins.Admin)

	for i := range conf.AdminConfs {
		adminConf := conf.AdminConfs[i]
		a, err := plugins.NewAdmin(&adminConf)
		if err != nil {
			fmt.Printf("cannot create admin plugin '%s': %v\n", adminConf.Name, err)
		}

		p.admins[adminConf.Name] = a
	}
}

func createApps(conf *configs.Project, p *Project) {
	p.apps = make(map[string]*App, len(conf.Apps))

	for _, appConf := range conf.Apps {
		appUrl, err := createAppUrl(&appConf)
		if err != nil {
			fmt.Printf("cannot parse app url in app '%s': %v\n",
				appConf.Name, err)
		}

		p.apps[appConf.Name] = &App{
			name:           appConf.Name,
			url:            appUrl,
			pathPrefix:     appConf.PathPrefix,
			authSessionExp: appConf.AuthSessionExp,
		}
	}
}

func createAppUrl(app *configs.App) (*url.URL, error) {
	if !strings.HasPrefix(app.Host, "http") {
		app.Host = "https://" + app.Host
	}

	appUrl, err := url.Parse(app.Host + app.PathPrefix)
	if err != nil {
		return nil, err
	}

	return appUrl, nil
}

func createAppPlugins(conf *configs.Project, p *Project) {
	for appName := range p.apps {
		appState := p.apps[appName]

		var appConf configs.App
		for _, a := range conf.Apps {
			if a.Name == appName {
				appConf = a
			}
		}

		appState.authenticators = createAuthenticators(&appConf)
		appState.identityManager = createIdentityManager(&appConf)

		var err error
		appState.authorizer, err = p.GetAuthorizer(appConf.Authz)
		if err != nil {
			fmt.Printf("app '%s': %v", appState.name, err)
		}
		if appConf.SecondFactor != "" {
			appState.secondFactor, err = p.GetSecondFactor(appConf.SecondFactor)
			if err != nil {
				fmt.Printf("app '%s': %v", appState.name, err)
			}
		}
	}
}

func createAuthenticators(app *configs.App) map[string]plugins.Authenticator {
	clearAuthnDuplicate(app)
	authenticators := make(map[string]plugins.Authenticator, len(app.Authn))

	for i := range app.Authn {
		authnConf := app.Authn[i]
		authenticator, err := plugins.NewAuthN(&authnConf)
		if err != nil {
			fmt.Printf("cannot create authenticator '%s' in app '%s': %v\n",
				authnConf.Type, app.Name, err)
		}
		authenticators[authnConf.Type] = authenticator
	}

	return authenticators
}

func clearAuthnDuplicate(app *configs.App) {
	for i := 0; i < len(app.Authn)-1; i++ {
		for j := i + 1; j < len(app.Authn); j++ {
			if app.Authn[i].Type == app.Authn[j].Type {
				copy(app.Authn[j:], app.Authn[j+1:])
				app.Authn = app.Authn[:len(app.Authn)-1]
			}
		}
	}
}

func createIdentityManager(app *configs.App) identity.ManagerI {
	i, err := identity.Create()
	if err != nil {
		fmt.Printf("cannot create idetity manager for app '%s'", app.Name)
	}
	return i
}

func initGlobalPlugins(p *Project) {
	initStorages(p)
	initKeyStorages(p)
	initPwHashers(p)
	initSenders(p)
	initCryptoKeys(p)
	initAdmins(p)
}

func initStorages(p *Project) {
	for name, s := range p.storages {
		if err := s.(PluginInitializer).Init(initAPI(p)); err != nil {
			fmt.Printf("cannot init storage '%s': %v\n", name, err)
			p.storages[name] = nil
		}
	}
}

func initKeyStorages(p *Project) {
	for name, s := range p.keyStorages {
		if err := s.(PluginInitializer).Init(initAPI(p)); err != nil {
			fmt.Printf("cannot init key storage '%s': %v\n", name, err)
			p.keyStorages[name] = nil
		}
	}
}

func initPwHashers(p *Project) {
	for name, h := range p.hashers {
		if err := h.(PluginInitializer).Init(initAPI(p)); err != nil {
			fmt.Printf("cannot init hasher '%s': %v\n", name, err)
			p.hashers[name] = nil
		}
	}
}

func initSenders(p *Project) {
	for name, s := range p.senders {
		if err := s.(PluginInitializer).Init(initAPI(p)); err != nil {
			fmt.Printf("cannot init sender '%s': %v\n", name, err)
			p.senders[name] = nil
		}
	}
}

func initCryptoKeys(p *Project) {
	for name, k := range p.cryptoKeys {
		if err := k.(PluginInitializer).Init(initAPI(p, withRouter(getRouter()))); err != nil {
			fmt.Printf("cannot init kstorage '%s': %v\n", name, err)
			p.cryptoKeys[name] = nil
		}
	}
}

func initAdmins(p *Project) {
	for name, a := range p.admins {
		if err := a.(PluginInitializer).Init(initAPI(p, withRouter(getRouter()))); err != nil {
			fmt.Printf("cannot init admin plugin '%s': %v\n", name, err)
			p.admins[name] = nil
		}
	}
}

func initAppPlugins(p *Project) {
	for _, a := range p.apps {
		initAuthenticators(a, p)
		initAuthorizer(a, p)
		initSecondFactor(a, p)
	}
}

func initAuthenticators(app *App, p *Project) {
	var routes []*Route

	for name, authenticator := range app.authenticators {
		prefix := fmt.Sprintf("%s$%s$", app.name, authenticator.GetMetaData().ID)
		pluginAPI := initAPI(p, withKeyPrefix(prefix), withApp(app), withRouter(getRouter()))

		if err := authenticator.(AppPluginInitializer).Init(app.name, pluginAPI); err != nil {
			fmt.Printf("cannot init authenticator '%s' in app '%s': %v\n",
				name, app.name, err)
			app.authenticators[name] = nil
		} else {
			pathPrefix := "/" + strings.ReplaceAll(authenticator.GetMetaData().Type, "_", "-")
			routes = append(routes, &Route{
				Method:  http.MethodPost,
				Path:    pathPrefix + "/login",
				Handler: handleLogin(authenticator.Login(), p, app),
			})
		}
	}
	getRouter().addAppRoutes(app.name, routes)
}

func initAuthorizer(app *App, p *Project) {
	prefix := fmt.Sprintf("%s$%s$", app.name, app.authorizer.GetMetaData().ID)
	pluginAPI := initAPI(p, withKeyPrefix(prefix), withApp(app), withRouter(getRouter()))

	if err := app.authorizer.(AppPluginInitializer).Init(app.name, pluginAPI); err != nil {
		fmt.Printf("cannot init authorizer in app '%s': %v\n", app.name, err)
		app.authorizer = nil
	}
}

func initSecondFactor(app *App, p *Project) {
	var routes []*Route

	if app.secondFactor != nil {
		prefix := fmt.Sprintf("%s$%s$", app.name, app.secondFactor.GetMetaData().ID)
		pluginAPI := initAPI(p, withKeyPrefix(prefix), withApp(app), withRouter(getRouter()))

		if err := app.secondFactor.(AppPluginInitializer).Init(app.name, pluginAPI); err != nil {
			fmt.Printf("cannot init second factor in app '%s': %v\n", app.name, err)
			app.secondFactor = nil
		} else {
			pathPrefix := "/2fa/" + strings.ReplaceAll(app.secondFactor.GetMetaData().Type, "_", "-")
			routes = append(routes, &Route{
				Method:  http.MethodPost,
				Path:    pathPrefix + "/verify",
				Handler: handle2FA(app.secondFactor.Verify(), p, app),
			})
		}
	}
	getRouter().addAppRoutes(app.name, routes)
}

func listPluginStatus() {
	fmt.Println("AUREOLE PLUGINS STATUS")

	for appName, app := range p.apps {
		fmt.Printf("\nAPP: %s\n", appName)

		printStatus("identity manager", p.apps[appName].identityManager)

		for name, authn := range app.authenticators {
			printStatus(name, authn)
		}
	}

	if len(p.authorizers) != 0 {
		fmt.Println("\nAUTHORIZERS")
		for name, plugin := range p.authorizers {
			printStatus(name, plugin)
		}
	}

	if len(p.secondFactors) != 0 {
		fmt.Println("\n2FA")
		for name, plugin := range p.secondFactors {
			printStatus(name, plugin)
		}
	}

	if len(p.storages) != 0 {
		fmt.Println("\nSTORAGE PLUGINS")
		for name, plugin := range p.storages {
			printStatus(name, plugin)
		}
	}

	if len(p.keyStorages) != 0 {
		fmt.Println("\nKEY STORAGE PLUGINS")
		for name, plugin := range p.keyStorages {
			printStatus(name, plugin)
		}
	}

	if len(p.hashers) != 0 {
		fmt.Println("\nHASHER PLUGINS")
		for name, plugin := range p.hashers {
			printStatus(name, plugin)
		}
	}

	if len(p.senders) != 0 {
		fmt.Println("\nSENDER PLUGINS")
		for name, plugin := range p.senders {
			printStatus(name, plugin)
		}
	}

	if len(p.cryptoKeys) != 0 {
		fmt.Println("\nCRYPTOKEY PLUGINS")
		for name, plugin := range p.cryptoKeys {
			printStatus(name, plugin)
		}
	}

	if len(p.admins) != 0 {
		fmt.Println("\nADMIN PLUGINS")
		for name, plugin := range p.admins {
			printStatus(name, plugin)
		}
	}
}

func printStatus(name string, plugin interface{}) {
	colorRed := "\033[31m"
	colorGreen := "\033[32m"
	resetColor := "\033[0m"

	checkMark, _ := utf8.DecodeRuneInString("\u2714")
	crossMark, _ := utf8.DecodeRuneInString("\u274c")

	if plugin != nil && !reflect.ValueOf(plugin).IsNil() {
		fmt.Printf("%s%s - %v%s\n", colorGreen, name, string(checkMark), resetColor)
	} else {
		fmt.Printf("%s%s - %v%s\n", colorRed, name, string(crossMark), resetColor)
	}
}
