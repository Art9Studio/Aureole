package core

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"unicode/utf8"
)

var p *project

type PluginInitializer interface {
	Init(api PluginAPI) error
}

var PluginInitErr = errors.New("plugin doesn't implement PluginInitializer interface")

func Init(conf *configs.Project) {
	p = &project{
		apiVersion: conf.APIVersion,
		testRun:    conf.TestRun,
		pingPath:   conf.PingPath,
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

	createApps(conf, p)
	initApps(p)
	listPluginStatus()
}

func createApps(conf *configs.Project, p *project) {
	p.apps = make(map[string]*app, len(conf.Apps))

	for _, appConf := range conf.Apps {
		appUrl, err := createAppUrl(appConf)
		if err != nil {
			fmt.Printf("cannot parse app url in app '%s': %v\n",
				appConf.Name, err)
		}
		app := &app{
			name:           appConf.Name,
			url:            appUrl,
			pathPrefix:     appConf.PathPrefix,
			authSessionExp: appConf.AuthSessionExp,
		}

		createPwHashers(app, appConf)
		createSenders(app, appConf)
		createCryptoKeys(app, appConf)
		createStorages(app, appConf)
		createKeyStorages(app, appConf)
		createAdmins(app, appConf)
		createAuthenticators(app, appConf)
		createAuthorizer(app, appConf)
		createSecondFactors(app, appConf)
		createIDManager(app, appConf)
		createAureoleService(app, appConf)

		p.apps[appConf.Name] = app
	}
}

func createAureoleService(app *app, conf configs.App) {
	var ok bool
	app.service.signKey, ok = app.getCryptoKey(conf.Service.SignKey)
	if !ok {
		fmt.Printf("app %s: cannot get service key\n", app.name)
	}
	app.service.encKey, ok = app.getCryptoKey(conf.Service.EncKey)
	if !ok {
		fmt.Printf("app %s: cannot get service key\n", app.name)
	}
	app.service.storage, ok = app.getStorage(conf.Service.Storage)
	if !ok {
		fmt.Printf("app %s: cannot get service storage\n", app.name)
	}
}

func createAppUrl(app configs.App) (*url.URL, error) {
	if !strings.HasPrefix(app.Host, "http") {
		app.Host = "https://" + app.Host
	}

	appUrl, err := url.Parse(app.Host + app.PathPrefix)
	if err != nil {
		return nil, err
	}

	return appUrl, nil
}

func createPwHashers(app *app, conf configs.App) {
	app.hashers = make(map[string]plugins.PWHasher)
	for i := range conf.HasherConfs {
		hasherConf := conf.HasherConfs[i]
		pwHasher, err := plugins.NewPWHasher(&conf.HasherConfs[i])
		if err != nil {
			fmt.Printf("app %s: cannot create hasher %s: %v\n", app.name, hasherConf.Name, err)
		}
		app.hashers[hasherConf.Name] = pwHasher
	}
}

func createSenders(app *app, conf configs.App) {
	app.senders = make(map[string]plugins.Sender)
	for i := range conf.Senders {
		senderConf := conf.Senders[i]
		sender, err := plugins.NewSender(&senderConf)
		if err != nil {
			fmt.Printf("app %s: cannot create sender %s: %v\n", app.name, senderConf.Name, err)
		}
		app.senders[senderConf.Name] = sender
	}
}

func createCryptoKeys(app *app, conf configs.App) {
	app.cryptoKeys = make(map[string]plugins.CryptoKey)
	for i := range conf.CryptoKeys {
		ckeyConf := conf.CryptoKeys[i]
		cryptoKey, err := plugins.NewCryptoKey(&ckeyConf)
		if err != nil {
			fmt.Printf("app %s: cannot create crypto key %s: %v\n", app.name, ckeyConf.Name, err)
		}
		app.cryptoKeys[ckeyConf.Name] = cryptoKey
	}
}

func createStorages(app *app, conf configs.App) {
	app.storages = make(map[string]plugins.Storage)
	for i := range conf.Storages {
		storageConf := conf.Storages[i]
		storage, err := plugins.NewStorage(&storageConf)
		if err != nil {
			fmt.Printf("app %s: cannot create storage %s: %v\n", app.name, storageConf.Name, err)
		}
		app.storages[storageConf.Name] = storage
	}
}

func createKeyStorages(app *app, conf configs.App) {
	app.keyStorages = make(map[string]plugins.KeyStorage)
	for i := range conf.KeyStorages {
		storageConf := conf.KeyStorages[i]
		keyStorage, err := plugins.NewKeyStorage(&storageConf)
		if err != nil {
			fmt.Printf("app %s: cannot create key storage %s: %v\n", app.name, storageConf.Name, err)
		}
		app.keyStorages[storageConf.Name] = keyStorage
	}
}

func createAdmins(app *app, conf configs.App) {
	app.admins = make(map[string]plugins.Admin)
	for i := range conf.AdminConfs {
		adminConf := conf.AdminConfs[i]
		admin, err := plugins.NewAdmin(&adminConf)
		if err != nil {
			fmt.Printf("app %s: cannot create admin plugin %s: %v\n", app.name, adminConf.Name, err)
		}
		app.admins[adminConf.Name] = admin
	}
}

func createAuthenticators(app *app, conf configs.App) {
	clearAuthnDuplicate(&conf)

	app.authenticators = make(map[string]plugins.Authenticator, len(conf.Authn))
	for i := range conf.Authn {
		authnConf := conf.Authn[i]
		authenticator, err := plugins.NewAuthN(&authnConf)
		if err != nil {
			fmt.Printf("app %s: cannot create authenticator %s: %v\n", app.name, authnConf.Type, err)
		}
		app.authenticators[authnConf.Type] = authenticator
	}
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

func createAuthorizer(app *app, conf configs.App) {
	authorizer, err := plugins.NewAuthZ(&conf.Authz)
	if err != nil {
		fmt.Printf("app %s: cannot create authorizator: %v\n", app.name, err)
	}
	app.authorizer = authorizer
}

func createSecondFactors(app *app, conf configs.App) {
	app.secondFactors = make(map[string]plugins.SecondFactor)
	for i := range conf.SecondFactors {
		mfaConf := conf.SecondFactors[i]
		secondFactor, err := plugins.NewSecondFactor(&mfaConf)
		if err != nil {
			fmt.Printf("app %s: cannot create second factor %s: %v\n", app.name, mfaConf.Name, err)
		}
		app.secondFactors[mfaConf.Name] = secondFactor
	}
}

func createIDManager(app *app, conf configs.App) {
	idManager, err := plugins.NewIDManager(&conf.IDManager)
	if err != nil {
		fmt.Printf("app %s: cannot create identity manager: %v\n", app.name, err)
	}
	app.idManager = idManager
}

func initApps(p *project) {
	for _, app := range p.apps {
		initStorages(app, p)
		initKeyStorages(app, p)
		initPwHashers(app, p)
		initSenders(app, p)
		initCryptoKeys(app, p)
		initAdmins(app, p)
		initIDManager(app, p)
		initAuthorizer(app, p)
		initSecondFactor(app, p)
		initAuthenticators(app, p)
	}
}

func initStorages(app *app, p *project) {
	for name, s := range app.storages {
		pluginInit, ok := s.(PluginInitializer)
		if ok {
			err := pluginInit.Init(initAPI(withProject(p), withApp(app)))
			if err != nil {
				fmt.Printf("app %s: cannot init storage '%s': %v\n", app.name, name, err)
				app.storages[name] = nil
			}
		} else {
			fmt.Printf("app %s: cannot init storage '%s': %v\n", app.name, name, PluginInitErr)
			app.storages[name] = nil
		}
	}
}

func initKeyStorages(app *app, p *project) {
	for name, s := range app.keyStorages {
		pluginInit, ok := s.(PluginInitializer)
		if ok {
			err := pluginInit.Init(initAPI(withProject(p), withApp(app)))
			if err != nil {
				fmt.Printf("app %s: cannot init key storage '%s': %v\n", app.name, name, err)
				app.keyStorages[name] = nil
			}
		} else {
			fmt.Printf("app %s: cannot init key storage '%s': %v\n", app.name, name, PluginInitErr)
			app.keyStorages[name] = nil
		}
	}
}

func initPwHashers(app *app, p *project) {
	for name, h := range app.hashers {
		pluginInit, ok := h.(PluginInitializer)
		if ok {
			err := pluginInit.Init(initAPI(withProject(p), withApp(app)))
			if err != nil {
				fmt.Printf("app %s: cannot init hasher '%s': %v\n", app.name, name, err)
				app.hashers[name] = nil
			}
		} else {
			fmt.Printf("app %s: cannot init hasher '%s': %v\n", app.name, name, PluginInitErr)
			app.hashers[name] = nil
		}
	}
}

func initSenders(app *app, p *project) {
	for name, s := range app.senders {
		pluginInit, ok := s.(PluginInitializer)
		if ok {
			err := pluginInit.Init(initAPI(withProject(p), withApp(app)))
			if err != nil {
				fmt.Printf("app %s: cannot init sender '%s': %v\n", app.name, name, err)
				app.senders[name] = nil
			}
		} else {
			fmt.Printf("app %s: cannot init sender '%s': %v\n", app.name, name, PluginInitErr)
			app.senders[name] = nil
		}
	}
}

func initCryptoKeys(app *app, p *project) {
	for name, k := range app.cryptoKeys {
		pluginInit, ok := k.(PluginInitializer)
		if ok {
			err := pluginInit.Init(initAPI(withProject(p), withApp(app), withRouter(getRouter())))
			if err != nil {
				fmt.Printf("app %s: cannot init kstorage '%s': %v\n", app.name, name, err)
				app.cryptoKeys[name] = nil
			}
		} else {
			fmt.Printf("app %s: cannot init kstorage '%s': %v\n", app.name, name, PluginInitErr)
			app.cryptoKeys[name] = nil
		}
	}
}

func initAdmins(app *app, p *project) {
	for name, a := range app.admins {
		pluginInit, ok := a.(PluginInitializer)
		if ok {
			err := pluginInit.Init(initAPI(withProject(p), withApp(app), withRouter(getRouter())))
			if err != nil {
				fmt.Printf("app %s: cannot init admin plugin '%s': %v\n", app.name, name, err)
				app.admins[name] = nil
			}
		} else {
			fmt.Printf("app %s: cannot init admin plugin '%s': %v\n", app.name, name, PluginInitErr)
			app.admins[name] = nil
		}
	}
}

func initAuthenticators(app *app, p *project) {
	var routes []*Route

	for name, authenticator := range app.authenticators {
		pluginInit, ok := authenticator.(PluginInitializer)
		if ok {
			prefix := fmt.Sprintf("%s$%s$", app.name, authenticator.GetMetaData().ID)
			pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(getRouter()))
			err := pluginInit.Init(pluginAPI)
			if err != nil {
				fmt.Printf("app %s: cannot init authenticator %s: %v\n", app.name, name, err)
				app.authenticators[name] = nil
			} else {
				pathPrefix := "/" + strings.ReplaceAll(authenticator.GetMetaData().Type, "_", "-")
				routes = append(routes, &Route{
					Method:  http.MethodPost,
					Path:    pathPrefix + "/login",
					Handler: handleLogin(authenticator.Login(), app),
				})
			}
		} else {
			fmt.Printf("app %s: cannot init authenticator %s: %v\n", app.name, name, PluginInitErr)
			app.authenticators[name] = nil
		}
	}
	getRouter().addAppRoutes(app.name, routes)
}

func initAuthorizer(app *app, p *project) {
	pluginInit, ok := app.authorizer.(PluginInitializer)
	if ok {
		prefix := fmt.Sprintf("%s$%s$", app.name, app.authorizer.GetMetaData().ID)
		pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(getRouter()))
		err := pluginInit.Init(pluginAPI)
		if err != nil {
			fmt.Printf("app %s: cannot init authorizer: %v\n", app.name, err)
			app.authorizer = nil
		}
	} else {
		fmt.Printf("app %s: cannot init authorizer: %v\n", app.name, PluginInitErr)
		app.authorizer = nil
	}
}

func initSecondFactor(app *app, p *project) {
	var routes []*Route

	if app.secondFactors != nil && len(app.secondFactors) != 0 {
		for name := range app.secondFactors {
			secondFactor := app.secondFactors[name]
			pluginInit, ok := secondFactor.(PluginInitializer)
			if ok {
				prefix := fmt.Sprintf("%s$%s$", app.name, secondFactor.GetMetaData().ID)
				pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(getRouter()))

				err := pluginInit.Init(pluginAPI)
				if err != nil {
					fmt.Printf("app %s: cannot init second factor %s: %v\n", app.name, name, err)
					app.secondFactors[secondFactor.GetMetaData().Name] = nil
				} else {
					pathPrefix := "/2fa/" + strings.ReplaceAll(secondFactor.GetMetaData().Type, "_", "-")
					routes = append(routes,
						&Route{
							Method:  http.MethodPost,
							Path:    pathPrefix + "/start",
							Handler: handle2FAInit(secondFactor.Init2FA(), app),
						},
						&Route{
							Method:  http.MethodPost,
							Path:    pathPrefix + "/verify",
							Handler: handle2FAVerify(secondFactor.Verify(), app),
						})
					app.secondFactors[name] = secondFactor
				}
			} else {
				fmt.Printf("app %s: cannot init second factor %s: %v\n", app.name, name, PluginInitErr)
				app.secondFactors[secondFactor.GetMetaData().Name] = nil
			}
		}
	}
	getRouter().addAppRoutes(app.name, routes)
}

func initIDManager(app *app, p *project) {
	pluginInit, ok := app.idManager.(PluginInitializer)
	if ok {
		pluginAPI := initAPI(withProject(p), withApp(app), withRouter(getRouter()))
		err := pluginInit.Init(pluginAPI)
		if err != nil {
			fmt.Printf("app %s: cannot init id manager: %v\n", app.name, err)
			app.idManager = nil
		}
	} else {
		fmt.Printf("app %s: cannot init id manager: %v\n", app.name, PluginInitErr)
		app.idManager = nil
	}
}

func listPluginStatus() {
	fmt.Println("AUREOLE PLUGINS STATUS")

	for appName, app := range p.apps {
		fmt.Printf("\nAPP: %s\n", appName)

		printStatus("IDENTITY MANAGER", p.apps[appName].idManager)
		printStatus("AUTHORIZER", app.authorizer)

		fmt.Println("\nAUTHENTICATORS")
		for name, authn := range app.authenticators {
			printStatus(name, authn)
		}

		if len(app.secondFactors) != 0 {
			fmt.Println("\n2FA")
			for name, plugin := range app.secondFactors {
				printStatus(name, plugin)
			}
		}

		if len(app.storages) != 0 {
			fmt.Println("\nSTORAGE PLUGINS")
			for name, plugin := range app.storages {
				printStatus(name, plugin)
			}
		}

		if len(app.keyStorages) != 0 {
			fmt.Println("\nKEY STORAGE PLUGINS")
			for name, plugin := range app.keyStorages {
				printStatus(name, plugin)
			}
		}

		if len(app.hashers) != 0 {
			fmt.Println("\nHASHER PLUGINS")
			for name, plugin := range app.hashers {
				printStatus(name, plugin)
			}
		}

		if len(app.senders) != 0 {
			fmt.Println("\nSENDER PLUGINS")
			for name, plugin := range app.senders {
				printStatus(name, plugin)
			}
		}

		if len(app.cryptoKeys) != 0 {
			fmt.Println("\nCRYPTOKEY PLUGINS")
			for name, plugin := range app.cryptoKeys {
				printStatus(name, plugin)
			}
		}

		if len(app.admins) != 0 {
			fmt.Println("\nADMIN PLUGINS")
			for name, plugin := range app.admins {
				printStatus(name, plugin)
			}
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
