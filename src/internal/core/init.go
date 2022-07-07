package core

import (
	"aureole/internal/configs"
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/swaggo/swag"
	"github.com/xlab/treeprint"
)

type PluginInitializer interface {
	Init(api PluginAPI) error
}

var PluginInitErr = errors.New("Plugin doesn't implement PluginInitializer interface")

func InitProject(conf *configs.Project, r *router) *project {
	var p = &project{
		apiVersion: conf.APIVersion,
		testRun:    conf.TestRun,
		pingPath:   conf.PingPath,
	}

	if p.testRun {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	addHealthRoute(r)

	createApps(conf, p)
	initApps(p, r)
	listPluginStatus(conf, p)

	err := assembleDoc(p)
	if err != nil {
		fmt.Printf("cannot assemble swagger docs: %v", err)
	} else {
		swag.Register("swagger", &swagger{})
	}

	return p

}

func addHealthRoute(r *router) {
	r.addProjectRoutes([]*Route{
		{
			Method: http.MethodGet,
			Path:   "/health",
			Handler: func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"status": "OK"})
			},
		},
	})
}

func createApps(conf *configs.Project, p *project) {
	p.apps = make(map[string]*app, len(conf.Apps))
	var senderRepository = SenderRepo
	var cryptoKeyRepository = CryptoKeyRepo
	var storagesRepository = StorageRepo
	var cryptoStoragesRepository = CryptoStorageRepo
	var rootPluginsRepository = RootRepo
	var authenticatorsRepository = AuthenticatorRepo
	var issuerRepository = IssuerRepo
	var multiFactorsRepository = MFARepo
	var idmanagerRepository = IDManagerRepo

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

		createSenders(senderRepository, app, appConf)
		createCryptoKeys(cryptoKeyRepository, app, appConf)
		createStorages(storagesRepository, app, appConf)
		createCryptoStorages(cryptoStoragesRepository, app, appConf)
		createRootPlugins(rootPluginsRepository, app, appConf)
		createAuthenticators(authenticatorsRepository, app, appConf)
		createIssuer(issuerRepository, app, appConf)
		createMultiFactors(multiFactorsRepository, app, appConf)
		createIDManager(idmanagerRepository, app, appConf)
		createAureoleInternals(app, appConf)

		p.apps[appConf.Name] = app
	}
}

func createAureoleInternals(app *app, conf configs.App) {
	var ok bool
	app.internal.signKey, ok = app.getCryptoKey(conf.Internal.SignKey)
	if !ok {
		fmt.Printf("app %s: cannot get internal key\n", app.name)
	}
	app.internal.encKey, ok = app.getCryptoKey(conf.Internal.EncKey)
	if !ok {
		fmt.Printf("app %s: cannot get internal key\n", app.name)
	}
	app.internal.storage, ok = app.getStorage(conf.Internal.Storage)
	if !ok {
		panic(fmt.Sprintf("app %s: cannot get internal storage\n", app.name))
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

func createSenders(repository *Repository[Sender], app *app, conf configs.App) {
	app.senders = make(map[string]Sender)
	for i := range conf.Senders {
		senderConf := conf.Senders[i]
		creator := CreatePlugin[Sender]
		sender, err := creator(repository, senderConf, TypeSender)
		if err != nil {
			fmt.Printf("app %s: cannot create sender %s: %v\n", app.name, senderConf.Name, err)
		}
		app.senders[senderConf.Name] = sender
	}
}

func createCryptoKeys(repository *Repository[CryptoKey], app *app, conf configs.App) {
	app.cryptoKeys = make(map[string]CryptoKey)
	for i := range conf.CryptoKeys {
		ckeyConf := conf.CryptoKeys[i]
		creator := CreatePlugin[CryptoKey]
		cryptoKey, err := creator(repository, ckeyConf, TypeCryptoKey)
		if err != nil {
			fmt.Printf("app %s: cannot create crypto key %s, %v\n", app.name, ckeyConf.Name, err)
		}
		app.cryptoKeys[ckeyConf.Name] = cryptoKey
	}
}

func createStorages(repository *Repository[Storage], app *app, conf configs.App) {
	app.storages = make(map[string]Storage)
	for i := range conf.Storages {
		storageConf := conf.Storages[i]
		creator := CreatePlugin[Storage]
		storage, err := creator(repository, storageConf, TypeStorage)
		if err != nil {
			fmt.Printf("app %s: cannot create storage %s: %v\n", app.name, storageConf.Name, err)
		}
		app.storages[storageConf.Name] = storage
	}
}

func createCryptoStorages(repository *Repository[CryptoStorage], app *app, conf configs.App) {
	app.cryptoStorages = make(map[string]CryptoStorage)
	for i := range conf.CryptoStorages {
		storageConf := conf.CryptoStorages[i]
		creator := CreatePlugin[CryptoStorage]
		cryptoStorage, err := creator(repository, storageConf, TypeCryptoStorage)
		if err != nil {
			fmt.Printf("app %s: cannot create crypto storage %s: %v\n", app.name, storageConf.Name, err)
		}
		app.cryptoStorages[storageConf.Name] = cryptoStorage
	}
}

func createRootPlugins(repository *Repository[RootPlugin], app *app, conf configs.App) {
	app.rootPlugins = make(map[string]RootPlugin)
	for i := range conf.RootPlugins {
		rootPluginConf := conf.RootPlugins[i]
		creator := CreatePlugin[RootPlugin]
		rootPlugin, err := creator(repository, rootPluginConf, TypeRoot)
		if err != nil {
			fmt.Printf("app %s: cannot create rootPlugin Plugin %s: %v\n", app.name, rootPluginConf.Name, err)
		}
		app.rootPlugins[rootPluginConf.Name] = rootPlugin
	}
}

func createAuthenticators(repository *Repository[Authenticator], app *app, conf configs.App) {
	//clearAuthnDuplicate(&conf)

	app.authenticators = make(map[string]Authenticator, len(conf.Auth))
	for i := range conf.Auth {
		authnConf := conf.Auth[i]
		creator := CreatePlugin[Authenticator]
		authenticator, err := creator(repository, authnConf, TypeAuth)
		if err != nil {
			fmt.Printf("app %s: cannot create authenticator %s: %v\n", app.name, authnConf.Plugin, err)
		}
		app.authenticators[authnConf.Name] = authenticator
	}
}

func clearAuthnDuplicate(app *configs.App) {
	for i := 0; i < len(app.Auth)-1; i++ {
		for j := i + 1; j < len(app.Auth); j++ {
			if app.Auth[i].Plugin == app.Auth[j].Plugin {
				copy(app.Auth[j:], app.Auth[j+1:])
				app.Auth = app.Auth[:len(app.Auth)-1]
			}
		}
	}
}

func createIssuer(repository *Repository[Issuer], app *app, conf configs.App) {
	issuerConf := conf.Issuer
	creator := CreatePlugin[Issuer]
	issuer, err := creator(repository, issuerConf, TypeIssuer)

	if err != nil {
		fmt.Printf("app %s: cannot create issuer: %v\n", app.name, err)
	}
	app.issuer = issuer
}

func createMultiFactors(repository *Repository[MFA], app *app, conf configs.App) {
	app.mfa = make(map[string]MFA)
	for i := range conf.MFA {
		mfaConf := conf.MFA[i]
		creator := CreatePlugin[MFA]
		multiFactor, err := creator(repository, mfaConf, TypeMFA)
		if err != nil {
			fmt.Printf("app %s: cannot create second factor %s: %v\n", app.name, mfaConf.Name, err)
		}

		app.mfa[mfaConf.Name] = multiFactor
	}
}

func createIDManager(repository *Repository[IDManager], app *app, conf configs.App) {
	idManagerConf := conf.IDManager
	// IdManager is optional so skip if not set
	if idManagerConf.Plugin == "" {
		return
	}
	creator := CreatePlugin[IDManager]
	idManager, err := creator(repository, idManagerConf, TypeIDManager)
	if err != nil {
		fmt.Printf("app %s: cannot create identity manager: %v\n", app.name, err)
	}
	app.idManager = idManager
}

func initApps(p *project, r *router) {
	for _, app := range p.apps {
		initStorages(app, p)
		initCryptoStorages(app, p)
		initSenders(app, p)
		initCryptoKeys(app, p, r)
		initAdmins(app, p, r)
		initIDManager(app, p, r)
		initIssuer(app, p, r)
		initSecondFactor(app, p, r)
		initAuthenticators(app, p, r)

		err := isRSA(app.internal.encKey.GetPrivateSet())
		if err != nil {
			app.internal.encKey = nil
			fmt.Printf("app %s: internal key must be RSA key\n", app.name)
		}
	}
}

func isRSA(set jwk.Set) error {
	key, ok := set.Get(0)
	if !ok {
		return errors.New("cannot get internal key")
	}

	var rsaKey rsa.PrivateKey
	return key.Raw(&rsaKey)
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

func initCryptoStorages(app *app, p *project) {
	for name, s := range app.cryptoStorages {
		pluginInit, ok := s.(PluginInitializer)
		if ok {
			err := pluginInit.Init(initAPI(withProject(p), withApp(app)))
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

func initCryptoKeys(app *app, p *project, r *router) {
	for name, k := range app.cryptoKeys {
		pluginInit, ok := k.(PluginInitializer)
		if ok {
			err := pluginInit.Init(initAPI(withProject(p), withApp(app), withRouter(r)))
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

func initAdmins(app *app, p *project, r *router) {
	for name, a := range app.rootPlugins {
		pluginInit, ok := a.(PluginInitializer)
		if ok {
			err := pluginInit.Init(initAPI(withProject(p), withApp(app), withRouter(r)))
			if err != nil {
				fmt.Printf("app %s: cannot init admin Plugin '%s': %v\n", app.name, name, err)
				app.rootPlugins[name] = nil
			}
		} else {
			fmt.Printf("app %s: cannot init admin Plugin '%s': %v\n", app.name, name, PluginInitErr)
			app.rootPlugins[name] = nil
		}
	}
}

func initAuthenticators(app *app, p *project, r *router) {
	var routes []*Route

	for name, authenticator := range app.authenticators {
		pluginInit, ok := authenticator.(PluginInitializer)
		if ok {
			// todo: add runtime ID
			prefix := fmt.Sprintf("%s$%s$", app.name, authenticator.GetMetaData().Name)
			pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(r))
			err := pluginInit.Init(pluginAPI)
			if err != nil {
				fmt.Printf("app %s: cannot init authenticator %s: %v\n", app.name, name, err)
				app.authenticators[name] = nil
			} else {
				pathPrefix := "/" + strings.ReplaceAll(authenticator.GetMetaData().Name, "_", "-")
				authRoute := &Route{
					Method:  authenticator.GetLoginMethod(),
					Path:    pathPrefix + "/login",
					Handler: loginHandler(authenticator.GetLoginWrapper(), app),
				}
				if err != nil {
					fmt.Printf("app %s: cannot init authenticator %s: %v\n", app.name, name, err)
				}

				routes = append(routes, authRoute)
			}
		} else {
			fmt.Printf("app %s: cannot init authenticator %s: %v\n", app.name, name, PluginInitErr)
			app.authenticators[name] = nil
		}
	}
	r.addAppRoutes(app.name, routes)
}

func initIssuer(app *app, p *project, r *router) {
	pluginInit, ok := app.issuer.(PluginInitializer)
	if ok {
		// todo: add runtime ID
		prefix := fmt.Sprintf("%s$%s$", app.name, app.issuer.GetMetaData().Name)
		pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(r))
		err := pluginInit.Init(pluginAPI)
		if err != nil {
			fmt.Printf("app %s: cannot init issuer: %v\n", app.name, err)
			app.issuer = nil
		}
	} else {
		fmt.Printf("app %s: cannot init issuer: %v\n", app.name, PluginInitErr)
		app.issuer = nil
	}
}

func initSecondFactor(app *app, p *project, r *router) {
	var routes []*Route

	if app.mfa != nil && len(app.mfa) != 0 {
		for name := range app.mfa {
			secondFactor := app.mfa[name]
			pluginInit, ok := secondFactor.(PluginInitializer)
			if ok {
				// todo: add runtime ID
				prefix := fmt.Sprintf("%s$%s$", app.name, secondFactor.GetMetaData().Name)
				pluginAPI := initAPI(withProject(p), withKeyPrefix(prefix), withApp(app), withRouter(r))

				err := pluginInit.Init(pluginAPI)
				if err != nil {
					fmt.Printf("app %s: cannot init second factor %s: %v\n", app.name, name, err)
					app.mfa[secondFactor.GetMetaData().Name] = nil
				} else {
					pathPrefix := "/2fa/" + strings.ReplaceAll(string(secondFactor.GetMetaData().Name), "_", "-")
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
					app.mfa[name] = secondFactor
				}
			} else {
				fmt.Printf("app %s: cannot init second factor %s: %v\n", app.name, name, PluginInitErr)
				app.mfa[secondFactor.GetMetaData().Name] = nil
			}
		}
	}
	r.addAppRoutes(app.name, routes)
}

func initIDManager(app *app, p *project, r *router) {
	if app.idManager == nil {
		return
	}
	pluginInit, ok := app.idManager.(PluginInitializer)
	if ok {
		pluginAPI := initAPI(withProject(p), withApp(app), withRouter(r))
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

func listPluginStatus(conf *configs.Project, p *project) {
	tree := treeprint.New()
	tree.SetValue("PLUGIN STATUSES")

	for _, appConf := range conf.Apps {
		appName := appConf.Name
		app := p.apps[appName]

		branch := tree.AddBranch(fmt.Sprintf("APP \"%s\"", appName))
		if appConf.IDManager.Plugin != "" {
			branch.AddNode(formatStatus("IDENTITY MANAGER", p.apps[appName].idManager))
		}
		issuerPlugin := fmt.Sprintf("ISSUER (%s)", appConf.Issuer.Plugin)
		branch.AddNode(formatStatus(issuerPlugin, app.issuer))

		if len(appConf.Auth) != 0 {
			subbranch := branch.AddBranch("AUTHENTICATORS")
			for _, auth := range appConf.Auth {
				plugin := fmt.Sprintf("%s (%s)", auth.Name, auth.Plugin)
				subbranch.AddNode(formatStatus(plugin, app.authenticators[auth.Name]))
			}
		}

		if len(appConf.MFA) != 0 {
			subbranch := branch.AddBranch("MFA")
			for _, mfa := range appConf.MFA {
				plugin := fmt.Sprintf("%s (%s)", mfa.Name, mfa.Plugin)
				subbranch.AddNode(formatStatus(plugin, app.mfa[mfa.Name]))
			}
		}

		if len(appConf.Storages) != 0 {
			subbranch := branch.AddBranch("STORAGE PLUGINS")
			for _, storage := range appConf.Storages {
				plugin := fmt.Sprintf("%s (%s)", storage.Name, storage.Plugin)
				subbranch.AddNode(formatStatus(plugin, app.storages[storage.Name]))
			}
		}

		if len(appConf.CryptoStorages) != 0 {
			subbranch := branch.AddBranch("CRYPTO STORAGE PLUGINS")
			for _, storage := range appConf.CryptoStorages {
				plugin := fmt.Sprintf("%s (%s)", storage.Name, storage.Plugin)
				subbranch.AddNode(formatStatus(plugin, app.cryptoStorages[storage.Name]))
			}
		}

		if len(appConf.CryptoKeys) != 0 {
			subbranch := branch.AddBranch("CRYPTO KEYS PLUGINS")
			for _, key := range appConf.CryptoKeys {
				plugin := fmt.Sprintf("%s (%s)", key.Name, key.Plugin)
				subbranch.AddNode(formatStatus(plugin, app.cryptoKeys[key.Name]))
			}
		}

		if len(appConf.Senders) != 0 {
			subbranch := branch.AddBranch("SENDERS PLUGINS")
			for _, sender := range appConf.Senders {
				plugin := fmt.Sprintf("%s (%s)", sender.Name, sender.Plugin)
				subbranch.AddNode(formatStatus(plugin, app.senders[sender.Name]))
			}
		}

		if len(appConf.RootPlugins) != 0 {
			subbranch := branch.AddBranch("ROOT PLUGINS")
			for _, root := range appConf.RootPlugins {
				plugin := fmt.Sprintf("%s (%s)", root.Name, root.Plugin)
				subbranch.AddNode(formatStatus(plugin, app.rootPlugins[root.Name]))
			}
		}
	}

	fmt.Println()
	fmt.Println(tree.String())
}

func formatStatus(name string, plugin interface{}) string {
	colorRed := "\033[31m"
	colorGreen := "\033[32m"
	resetColor := "\033[0m"

	checkMark, _ := utf8.DecodeRuneInString("\u2714")
	crossMark, _ := utf8.DecodeRuneInString("\u274c")

	if plugin != nil && !reflect.ValueOf(plugin).IsNil() {
		return fmt.Sprintf("%s%s - %v%s", colorGreen, name, string(checkMark), resetColor)
	} else {
		return fmt.Sprintf("%s%s - %v%s", colorRed, name, string(crossMark), resetColor)
	}
}
