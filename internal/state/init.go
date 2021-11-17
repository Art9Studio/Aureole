package state

import (
	"aureole/internal/configs"
	"aureole/internal/identity"
	"aureole/internal/plugins/admin"
	adminT "aureole/internal/plugins/admin/types"
	"aureole/internal/plugins/authn"
	authnT "aureole/internal/plugins/authn/types"
	"aureole/internal/plugins/authz"
	authzT "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/cryptokey"
	cryptoKeyT "aureole/internal/plugins/cryptokey/types"
	"aureole/internal/plugins/kstorage"
	kstorageT "aureole/internal/plugins/kstorage/types"
	"aureole/internal/plugins/pwhasher"
	pwhasherT "aureole/internal/plugins/pwhasher/types"
	"aureole/internal/plugins/sender"
	senderT "aureole/internal/plugins/sender/types"
	"aureole/internal/plugins/storage"
	storageT "aureole/internal/plugins/storage/types"
	"aureole/internal/router"
	_interface "aureole/internal/router/interface"
	"aureole/internal/state/app"
	"crypto/tls"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/url"
	"strings"
)

func Init(conf *configs.Project, p *Project) {
	p.APIVersion = conf.APIVersion
	p.TestRun = conf.TestRun
	p.PingPath = conf.PingPath

	if p.TestRun {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	router.Router.AddProjectRoutes([]*_interface.Route{
		{
			Method: "GET",
			Path:   p.PingPath,
			Handler: func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			},
		},
	})

	createGlobalPlugins(conf, p)
	createApps(conf, p)
	createAppPlugins(conf, p)

	initGlobalPlugins(p)
	initAppPlugins(p)

	var err error
	if p.Service.internalKey, err = p.GetCryptoKey(conf.Service.InternalKey); err != nil {
		fmt.Printf("cannot init service key: %v\n", err)
	}
	if p.Service.storage, err = p.GetStorage(conf.Service.Storage); err != nil {
		fmt.Printf("cannot init service storage: %v\n", err)
	}
}

func createGlobalPlugins(conf *configs.Project, p *Project) {
	createPwHashers(conf, p)
	createSenders(conf, p)
	createCryptoKeys(conf, p)
	createStorages(conf, p)
	createKeyStorages(conf, p)
	createAdmins(conf, p)
}

func createPwHashers(conf *configs.Project, p *Project) {
	p.Hashers = make(map[string]pwhasherT.PwHasher)

	for i := range conf.HasherConfs {
		hasherConf := conf.HasherConfs[i]
		h, err := pwhasher.New(&conf.HasherConfs[i])
		if err != nil {
			fmt.Printf("cannot create hasher '%s': %v\n", hasherConf.Name, err)
		}

		p.Hashers[hasherConf.Name] = h
	}
}

func createSenders(conf *configs.Project, p *Project) {
	p.Senders = make(map[string]senderT.Sender)

	for i := range conf.Senders {
		senderConf := conf.Senders[i]
		s, err := sender.New(&senderConf)
		if err != nil {
			fmt.Printf("cannot create sender '%s': %v\n", senderConf.Name, err)
		}

		p.Senders[senderConf.Name] = s
	}
}

func createCryptoKeys(conf *configs.Project, p *Project) {
	p.CryptoKeys = make(map[string]cryptoKeyT.CryptoKey)

	for i := range conf.CryptoKeys {
		ckeyConf := conf.CryptoKeys[i]
		ckey, err := cryptokey.New(&ckeyConf)
		if err != nil {
			fmt.Printf("cannot create crypto key '%s': %v\n", ckeyConf.Name, err)
		}

		p.CryptoKeys[ckeyConf.Name] = ckey
	}
}

func createStorages(conf *configs.Project, p *Project) {
	p.Storages = make(map[string]storageT.Storage)

	for i := range conf.Storages {
		storageConf := conf.Storages[i]
		connSess, err := storage.New(&storageConf)
		if err != nil {
			fmt.Printf("open connection session to storage '%s': %v\n", storageConf.Name, err)
		}

		p.Storages[storageConf.Name] = connSess
	}
}

func createKeyStorages(conf *configs.Project, p *Project) {
	p.KeyStorages = make(map[string]kstorageT.KeyStorage)

	for i := range conf.KeyStorages {
		storageConf := conf.KeyStorages[i]
		connSess, err := kstorage.New(&storageConf)
		if err != nil {
			fmt.Printf("open connection session to key storage '%s': %v\n", storageConf.Name, err)
		}

		p.KeyStorages[storageConf.Name] = connSess
	}
}

func createAdmins(conf *configs.Project, p *Project) {
	p.Admins = make(map[string]adminT.Admin)

	for i := range conf.AdminConfs {
		adminConf := conf.AdminConfs[i]
		a, err := admin.New(&adminConf)
		if err != nil {
			fmt.Printf("cannot create admin plugin '%s': %v\n", adminConf.Name, err)
		}

		p.Admins[adminConf.Name] = a
	}
}

func createApps(conf *configs.Project, p *Project) {
	p.Apps = make(map[string]*app.App, len(conf.Apps))

	for _, appConf := range conf.Apps {
		appUrl, err := createAppUrl(&appConf)
		if err != nil {
			fmt.Printf("cannot parse app url in app '%s': %v\n",
				appConf.Name, err)
		}

		p.Apps[appConf.Name] = &app.App{
			Name:       appConf.Name,
			Url:        appUrl,
			PathPrefix: appConf.PathPrefix,
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
	for appName := range p.Apps {
		appState := p.Apps[appName]

		var appConf configs.App
		for _, a := range conf.Apps {
			if a.Name == appName {
				appConf = a
			}
		}

		appState.Authenticators = createAuthenticators(&appConf)
		appState.Authorizer = createAuthorizer(&appConf)
		appState.IdentityManager = createIdentityManager(&appConf)
	}
}

func createAuthenticators(app *configs.App) map[string]authnT.Authenticator {
	authenticators := make(map[string]authnT.Authenticator, len(app.Authn))

	for i := range app.Authn {
		authnConf := app.Authn[i]
		authenticator, err := authn.New(&authnConf)
		if err != nil {
			fmt.Printf("cannot create authenticator '%s' in app '%s': %v\n",
				authnConf.Type, app.Name, err)
		}

		authenticators[authnConf.Type] = authenticator
	}

	return authenticators
}

func createAuthorizer(app *configs.App) authzT.Authorizer {
	authorizer, err := authz.New(&app.Authz)
	if err != nil {
		fmt.Printf("cannot create authorizer in app '%s': %v\n", app.Name, err)
	}

	return authorizer
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
	for name, s := range p.Storages {
		if err := s.Init(); err != nil {
			fmt.Printf("cannot init storage '%s': %v\n", name, err)
			p.Storages[name] = nil
		}
	}
}

func initKeyStorages(p *Project) {
	for name, s := range p.KeyStorages {
		if err := s.Init(); err != nil {
			fmt.Printf("cannot init key storage '%s': %v\n", name, err)
			p.KeyStorages[name] = nil
		}
	}
}

func initPwHashers(p *Project) {
	for name, h := range p.Hashers {
		if err := h.Init(); err != nil {
			fmt.Printf("cannot init hasher '%s': %v\n", name, err)
			p.Hashers[name] = nil
		}
	}
}

func initSenders(p *Project) {
	for name, s := range p.Senders {
		if err := s.Init(); err != nil {
			fmt.Printf("cannot init sender '%s': %v\n", name, err)
			p.Senders[name] = nil
		}
	}
}

func initCryptoKeys(p *Project) {
	for name, k := range p.CryptoKeys {
		if err := k.Init(); err != nil {
			fmt.Printf("cannot init kstorage '%s': %v\n", name, err)
			p.CryptoKeys[name] = nil
		}
	}
}

func initAdmins(p *Project) {
	for name, a := range p.Admins {
		if err := a.Init(); err != nil {
			fmt.Printf("cannot init admin plugin '%s': %v\n", name, err)
			p.Admins[name] = nil
		}
	}
}

func initAppPlugins(p *Project) {
	for name, a := range p.Apps {
		initAuthenticators(a)
		initAuthorizer(name, a)
	}
}

func initAuthenticators(app *app.App) {
	for name, authenticator := range app.Authenticators {
		if err := authenticator.Init(app); err != nil {
			fmt.Printf("cannot init authenticator '%s' in app '%s': %v\n",
				name, app.Name, err)
			app.Authenticators[name] = nil
		}
	}
}

func initAuthorizer(appName string, app *app.App) {
	if err := app.Authorizer.Init(appName); err != nil {
		fmt.Printf("cannot init authorizer in app '%s': %v\n", app.Name, err)
		app.Authorizer = nil
	}
}
