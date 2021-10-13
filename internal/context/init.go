package context

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	"aureole/internal/context/app"
	"aureole/internal/identity"
	"aureole/internal/plugins/admin"
	adminTypes "aureole/internal/plugins/admin/types"
	"aureole/internal/plugins/authn"
	authnTypes "aureole/internal/plugins/authn/types"
	"aureole/internal/plugins/authz"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/cryptokey"
	cryptoKeyTypes "aureole/internal/plugins/cryptokey/types"
	"aureole/internal/plugins/pwhasher"
	pwhasherTypes "aureole/internal/plugins/pwhasher/types"
	"aureole/internal/plugins/sender"
	senderTypes "aureole/internal/plugins/sender/types"
	"aureole/internal/plugins/storage"
	"aureole/internal/plugins/storage/types"
	"aureole/internal/router"
	_interface "aureole/internal/router/interface"
	"crypto/tls"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/url"
	"strings"
)

func Init(conf *configs.Project, ctx *ProjectCtx) {
	ctx.APIVersion = conf.APIVersion
	ctx.TestRun = conf.TestRun
	ctx.PingPath = conf.PingPath

	if ctx.TestRun {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	router.Router.AddProjectRoutes([]*_interface.Route{
		{
			Method: "GET",
			Path:   ctx.PingPath,
			Handler: func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			},
		},
	})

	createGlobalPlugins(conf, ctx)
	createApps(conf, ctx)
	createAppPlugins(conf, ctx)

	createCollections(conf, ctx)
	createIdentities(conf, ctx)

	initGlobalPlugins(ctx)
	initAppPlugins(ctx)

	initCollections(ctx)
}

func createGlobalPlugins(conf *configs.Project, ctx *ProjectCtx) {
	createPwHashers(conf, ctx)
	createSenders(conf, ctx)
	createCryptoKeys(conf, ctx)
	createStorages(conf, ctx)
	createAdmins(conf, ctx)
}

func createAppPlugins(conf *configs.Project, ctx *ProjectCtx) {
	for appName := range ctx.Apps {
		appCtx := ctx.Apps[appName]

		var appConf configs.App
		for _, a := range conf.Apps {
			if a.Name == appName {
				appConf = a
			}
		}

		appCtx.Authenticators = createAuthenticators(&appConf)
		appCtx.Authorizers = createAuthorizers(&appConf)
	}
}

func createStorages(conf *configs.Project, ctx *ProjectCtx) {
	ctx.Storages = make(map[string]types.Storage)

	for i := range conf.StorageConfs {
		storageConf := conf.StorageConfs[i]
		connSess, err := storage.New(&storageConf)
		if err != nil {
			fmt.Printf("open connection session to storage '%s': %v\n", storageConf.Name, err)
		}

		ctx.Storages[storageConf.Name] = connSess
	}

	cleanupStorages(conf, ctx)
}

func cleanupStorages(conf *configs.Project, ctx *ProjectCtx) {
	isUsedStorage := make(map[string]bool)

	for storageName := range ctx.Storages {
		isUsedStorage[storageName] = false

		for _, appConf := range conf.Apps {
			for _, authzItem := range appConf.Authz {
				if storageName == authzItem.Config["storage"] {
					isUsedStorage[storageName] = true
					break
				}
			}

			for _, authnItem := range appConf.Authn {
				if storageName == authnItem.Config["storage"] {
					isUsedStorage[storageName] = true
					break
				}
			}
		}
	}
}

func createAdmins(conf *configs.Project, ctx *ProjectCtx) {
	ctx.Admins = make(map[string]adminTypes.Admin)

	for i := range conf.AdminConfs {
		adminConf := conf.AdminConfs[i]
		a, err := admin.New(&adminConf)
		if err != nil {
			fmt.Printf("cannot create admin plugin '%s': %v\n", adminConf.Name, err)
		}

		ctx.Admins[adminConf.Name] = a
	}

	cleanupStorages(conf, ctx)
}

func createCollections(conf *configs.Project, ctx *ProjectCtx) {
	ctx.Collections = make(map[string]*collections.Collection)

	for _, collConf := range conf.CollConfs {
		coll, err := collections.Create(&collConf)
		if err != nil {
			fmt.Printf("cannot create collection '%s': %v\n", coll.Name, err)
		}
		ctx.Collections[collConf.Name] = coll
	}
}

func createPwHashers(conf *configs.Project, ctx *ProjectCtx) {
	ctx.Hashers = make(map[string]pwhasherTypes.PwHasher)

	for i := range conf.HasherConfs {
		hasherConf := conf.HasherConfs[i]
		h, err := pwhasher.New(&conf.HasherConfs[i])
		if err != nil {
			fmt.Printf("cannot create hasher '%s': %v\n", hasherConf.Name, err)
		}

		ctx.Hashers[hasherConf.Name] = h
	}
}

func createSenders(conf *configs.Project, ctx *ProjectCtx) {
	ctx.Senders = make(map[string]senderTypes.Sender)

	for i := range conf.Senders {
		senderConf := conf.Senders[i]
		s, err := sender.New(&senderConf)
		if err != nil {
			fmt.Printf("cannot create sender '%s': %v\n", senderConf.Name, err)
		}

		ctx.Senders[senderConf.Name] = s
	}
}

func createCryptoKeys(conf *configs.Project, ctx *ProjectCtx) {
	ctx.CryptoKeys = make(map[string]cryptoKeyTypes.CryptoKey)

	for i := range conf.CryptoKeys {
		ckeyConf := conf.CryptoKeys[i]
		ckey, err := cryptokey.New(&ckeyConf)
		if err != nil {
			fmt.Printf("cannot create crypto key '%s': %v\n", ckeyConf.Name, err)
		}

		ctx.CryptoKeys[ckeyConf.Name] = ckey
	}
}

func createApps(conf *configs.Project, ctx *ProjectCtx) {
	ctx.Apps = make(map[string]*app.App, len(conf.Apps))

	for _, appConf := range conf.Apps {
		appUrl, err := createAppUrl(&appConf)
		if err != nil {
			fmt.Printf("cannot parse app url in app '%s': %v\n",
				appConf.Name, err)
		}

		ctx.Apps[appConf.Name] = &app.App{
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

func createIdentities(conf *configs.Project, ctx *ProjectCtx) {
	for appName := range ctx.Apps {
		appCtx := ctx.Apps[appName]

		var appConf configs.App
		for _, a := range conf.Apps {
			if a.Name == appName {
				appConf = a
			}
		}

		i, err := identity.Create(&appConf.Identity, ctx.Collections)
		if err != nil {
			fmt.Printf("cannot create idetity for app '%s'", appName)
			appCtx.Identity = nil
		} else {
			appCtx.Identity = i
		}
	}
}

func createAuthenticators(app *configs.App) map[string]authnTypes.Authenticator {
	authenticators := make(map[string]authnTypes.Authenticator, len(app.Authn))

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

func createAuthorizers(app *configs.App) map[string]authzTypes.Authorizer {
	authorizers := make(map[string]authzTypes.Authorizer, len(app.Authz))

	for i := range app.Authz {
		authzConf := app.Authz[i]
		authorizer, err := authz.New(&authzConf)
		if err != nil {
			fmt.Printf("cannot create authorizer '%s' in app '%s': %v\n",
				authzConf.Name, app.Name, err)
		}

		authorizers[authzConf.Name] = authorizer
	}

	return authorizers
}

func initCollections(ctx *ProjectCtx) {
	for collName := range ctx.Collections {
		coll := ctx.Collections[collName]
		err := coll.Init(ctx.Collections)
		if err != nil {
			fmt.Printf("cannot init collection '%s': %v\n", collName, err)
			ctx.Collections[collName] = nil
		}
	}
}

func initStorages(ctx *ProjectCtx) {
	for name, s := range ctx.Storages {
		if err := s.Init(); err != nil {
			fmt.Printf("cannot init storage '%s': %v\n", name, err)
			ctx.Storages[name] = nil
		} else if err := s.Ping(); err != nil {
			fmt.Printf("cannot ping storage '%s': %v\n", name, err)
			ctx.Storages[name] = nil
		}
	}
}

func initPwHashers(ctx *ProjectCtx) {
	for name, h := range ctx.Hashers {
		if err := h.Init(); err != nil {
			fmt.Printf("cannot init hasher '%s': %v\n", name, err)
			ctx.Hashers[name] = nil
		}
	}
}

func initSenders(ctx *ProjectCtx) {
	for name, s := range ctx.Senders {
		if err := s.Init(); err != nil {
			fmt.Printf("cannot init sender '%s': %v\n", name, err)
			ctx.Senders[name] = nil
		}
	}
}

func initCryptoKeys(ctx *ProjectCtx) {
	for name, k := range ctx.CryptoKeys {
		if err := k.Init(); err != nil {
			fmt.Printf("cannot init storage '%s': %v\n", name, err)
			ctx.CryptoKeys[name] = nil
		}
	}
}

func initAdmins(ctx *ProjectCtx) {
	for name, a := range ctx.Admins {
		if err := a.Init(); err != nil {
			fmt.Printf("cannot init admin plugin '%s': %v\n", name, err)
			ctx.Admins[name] = nil
		}
	}
}

func initGlobalPlugins(ctx *ProjectCtx) {
	initStorages(ctx)
	initPwHashers(ctx)
	initSenders(ctx)
	initCryptoKeys(ctx)
	initAdmins(ctx)
}

func initAppPlugins(ctx *ProjectCtx) {
	for name, a := range ctx.Apps {
		initAuthenticators(a)
		initAuthorizers(name, a)
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

func initAuthorizers(appName string, app *app.App) {
	for name, authorizer := range app.Authorizers {
		if err := authorizer.Init(appName); err != nil {
			fmt.Printf("cannot init authorizer '%s' in app '%s': %v\n",
				name, app.Name, err)
			app.Authorizers[name] = nil
		}
	}
}
