package core

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwk"
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
		router: &router{
			appRoutes:     map[string][]*Route{},
			projectRoutes: []*Route{},
			staticPaths:   map[string][]string{},
		},
	}

	if p.testRun {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	p.router.addProjectRoutes([]*Route{
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

	err := assembleSwagger()
	if err != nil {
		fmt.Printf("cannot assemble swagger docs: %v", err)
	}
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

		createSenders(app, appConf)
		createCryptoKeys(app, appConf)
		createStorages(app, appConf)
		createCryptoStorages(app, appConf)
		createAdmins(app, appConf)
		createAuthenticators(app, appConf)
		createAuthorizer(app, appConf)
		createSecondFactors(app, appConf)
		createIDManager(app, appConf)
		createAureoleService(app, appConf)
		createUI(app, appConf)

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

func createCryptoStorages(app *app, conf configs.App) {
	app.cryptoStorages = make(map[string]plugins.CryptoStorage)
	for i := range conf.CryptoStorages {
		storageConf := conf.CryptoStorages[i]
		cryptoStorage, err := plugins.NewCryptoStorage(&storageConf)
		if err != nil {
			fmt.Printf("app %s: cannot create crypto storage %s: %v\n", app.name, storageConf.Name, err)
		}
		app.cryptoStorages[storageConf.Name] = cryptoStorage
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

func createUI(app *app, conf configs.App) {
	ui, err := plugins.NewUI(&conf.UI)
	if err != nil {
		fmt.Printf("app %s: cannot create ui: %v\n", app.name, err)
	}
	app.ui = ui
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
		initCryptoStorages(app, p)
		initSenders(app, p)
		initCryptoKeys(app, p)
		initAdmins(app, p)
		initIDManager(app, p)
		initAuthorizer(app, p)
		initSecondFactor(app, p)
		initAuthenticators(app, p)
		initUI(app, p)

		err := isRSA(app.service.encKey.GetPrivateSet())
		if err != nil {
			app.service.encKey = nil
			fmt.Printf("app %s: service key must be RSA key\n", app.name)
		}
	}
}

func isRSA(set jwk.Set) error {
	key, ok := set.Get(0)
	if !ok {
		return errors.New("cannot get service key")
	}

	var rsaKey rsa.PrivateKey
	return key.Raw(&rsaKey)
}

func initStorages(app *app, p *project) {
	for name, storage := range app.storages {
		pluginInit, ok := storage.(PluginInitializer)
		if ok {
			prefix := fmt.Sprintf("%s$%s$", app.name, storage.GetMetaData().ID)
			pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(p.router))

			err := pluginInit.Init(pluginAPI)
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

func initCryptoStorages(app *app, p *project) {
	for name, cryptoStorage := range app.cryptoStorages {
		pluginInit, ok := cryptoStorage.(PluginInitializer)
		if ok {
			prefix := fmt.Sprintf("%s$%s$", app.name, cryptoStorage.GetMetaData().ID)
			pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(p.router))

			err := pluginInit.Init(pluginAPI)
			if err != nil {
				fmt.Printf("app %s: cannot init key storage '%s': %v\n", app.name, name, err)
				app.cryptoStorages[name] = nil
			}
		} else {
			fmt.Printf("app %s: cannot init key storage '%s': %v\n", app.name, name, PluginInitErr)
			app.cryptoStorages[name] = nil
		}
	}
}

func initSenders(app *app, p *project) {
	for name, sender := range app.senders {
		pluginInit, ok := sender.(PluginInitializer)
		if ok {
			prefix := fmt.Sprintf("%s$%s$", app.name, sender.GetMetaData().ID)
			pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(p.router))

			err := pluginInit.Init(pluginAPI)
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
	for name, cryptoKey := range app.cryptoKeys {
		pluginInit, ok := cryptoKey.(PluginInitializer)
		if ok {
			prefix := fmt.Sprintf("%s$%s$", app.name, cryptoKey.GetMetaData().ID)
			pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(p.router))

			err := pluginInit.Init(pluginAPI)
			if err != nil {
				fmt.Printf("app %s: cannot init crypto key '%s': %v\n", app.name, name, err)
				app.cryptoKeys[name] = nil
			}
		} else {
			fmt.Printf("app %s: cannot init crypto key '%s': %v\n", app.name, name, PluginInitErr)
			app.cryptoKeys[name] = nil
		}
	}
}

func initAdmins(app *app, p *project) {
	for name, admin := range app.admins {
		pluginInit, ok := admin.(PluginInitializer)
		if ok {
			prefix := fmt.Sprintf("%s$%s$", app.name, admin.GetMetaData().ID)
			pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(p.router))

			err := pluginInit.Init(pluginAPI)
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
			pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(p.router))

			err := pluginInit.Init(pluginAPI)
			if err != nil {
				fmt.Printf("app %s: cannot init authenticator %s: %v\n", app.name, name, err)
				app.authenticators[name] = nil
			} else {
				pathPrefix := "/" + strings.ReplaceAll(authenticator.GetMetaData().Type, "_", "-")
				specs, _ := authenticator.GetHandlersSpec()
				if err != nil {
					fmt.Printf("app %s: cannot init authenticator %s: %v\n", app.name, name, err)
				}

				var method string
				if specs.Paths["/login"].Post != nil {
					method = http.MethodPost
				} else {
					method = http.MethodGet
				}

				routes = append(routes, &Route{
					Method:  method,
					Path:    pathPrefix + "/login",
					Handler: loginHandler(authenticator.LoginWrapper(), app),
				})
			}
		} else {
			fmt.Printf("app %s: cannot init authenticator %s: %v\n", app.name, name, PluginInitErr)
			app.authenticators[name] = nil
		}
	}
	p.router.addAppRoutes(app.name, routes)
}

func initAuthorizer(app *app, p *project) {
	pluginInit, ok := app.authorizer.(PluginInitializer)
	if ok {
		prefix := fmt.Sprintf("%s$%s$", app.name, app.authorizer.GetMetaData().ID)
		pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(p.router))

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
				pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(p.router))

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
							Handler: mfaInitHandler(secondFactor.Init2FA(), app),
						},
						&Route{
							Method:  http.MethodPost,
							Path:    pathPrefix + "/verify",
							Handler: mfaVerificationHandler(secondFactor.Verify(), app),
						})
					app.secondFactors[name] = secondFactor
				}
			} else {
				fmt.Printf("app %s: cannot init second factor %s: %v\n", app.name, name, PluginInitErr)
				app.secondFactors[secondFactor.GetMetaData().Name] = nil
			}
		}
	}
	p.router.addAppRoutes(app.name, routes)
}

func initIDManager(app *app, p *project) {
	pluginInit, ok := app.idManager.(PluginInitializer)
	if ok {
		prefix := fmt.Sprintf("%s$%s$", app.name, app.idManager.GetMetaData().ID)
		pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(p.router))

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

func initUI(app *app, p *project) {
	pluginInit, ok := app.ui.(PluginInitializer)
	if ok {
		prefix := fmt.Sprintf("%s$%s$", app.name, app.ui.GetMetaData().ID)
		pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(p.router))

		err := pluginInit.Init(pluginAPI)
		if err != nil {
			fmt.Printf("app %s: cannot init ui: %v\n", app.name, err)
			app.ui = nil
		}
	} else {
		fmt.Printf("app %s: cannot init ui: %v\n", app.name, PluginInitErr)
		app.ui = nil
	}
}

func listPluginStatus() {
	fmt.Println("AUREOLE PLUGINS STATUS")

	for appName, app := range p.apps {
		fmt.Printf("\nAPP: %s\n", appName)

		printStatus("IDENTITY MANAGER", getPluginStatus(p.apps[appName].idManager))
		printStatus("AUTHORIZER", getPluginStatus(app.authorizer))

		fmt.Println("\nAUTHENTICATORS")
		for name, plugin := range app.authenticators {
			status := getPluginStatus(plugin)
			if !status {
				delete(app.authenticators, name)
			}
			printStatus(name, status)
		}

		if len(app.secondFactors) != 0 {
			fmt.Println("\n2FA")
			for name, plugin := range app.secondFactors {
				status := getPluginStatus(plugin)
				if !status {
					delete(app.secondFactors, name)
				}
				printStatus(name, status)
			}
		}

		if len(app.storages) != 0 {
			fmt.Println("\nSTORAGE PLUGINS")
			for name, plugin := range app.storages {
				status := getPluginStatus(plugin)
				if !status {
					delete(app.storages, name)
				}
				printStatus(name, status)
			}
		}

		if len(app.cryptoStorages) != 0 {
			fmt.Println("\nKEY STORAGE PLUGINS")
			for name, plugin := range app.cryptoStorages {
				status := getPluginStatus(plugin)
				if !status {
					delete(app.cryptoStorages, name)
				}
				printStatus(name, status)
			}
		}

		if len(app.senders) != 0 {
			fmt.Println("\nSENDER PLUGINS")
			for name, plugin := range app.senders {
				status := getPluginStatus(plugin)
				if !status {
					delete(app.senders, name)
				}
				printStatus(name, status)
			}
		}

		if len(app.cryptoKeys) != 0 {
			fmt.Println("\nCRYPTOKEY PLUGINS")
			for name, plugin := range app.cryptoKeys {
				status := getPluginStatus(plugin)
				if !status {
					delete(app.cryptoKeys, name)
				}
				printStatus(name, status)
			}
		}

		if len(app.admins) != 0 {
			fmt.Println("\nADMIN PLUGINS")
			for name, plugin := range app.admins {
				status := getPluginStatus(plugin)
				if !status {
					delete(app.admins, name)
				}
				printStatus(name, status)
			}
		}
	}
}

func getPluginStatus(plugin interface{}) bool {
	return plugin != nil && !reflect.ValueOf(plugin).IsNil()
}

func printStatus(name string, status bool) {
	colorRed := "\033[31m"
	colorGreen := "\033[32m"
	resetColor := "\033[0m"

	checkMark, _ := utf8.DecodeRuneInString("\u2714")
	crossMark, _ := utf8.DecodeRuneInString("\u274c")

	if status {
		fmt.Printf("%s%s - %v%s\n", colorGreen, name, string(checkMark), resetColor)
	} else {
		fmt.Printf("%s%s - %v%s\n", colorRed, name, string(crossMark), resetColor)
	}
}
