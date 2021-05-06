package context

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	"aureole/internal/context/app"
	"aureole/internal/identity"
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
	"fmt"
)

func Init(conf *configs.Project, ctx *ProjectCtx) error {
	ctx.APIVersion = conf.APIVersion

	if err := createGlobalPlugins(conf, ctx); err != nil {
		return err
	}
	if err := createApps(conf, ctx); err != nil {
		return err
	}
	if err := createAppPlugins(conf, ctx); err != nil {
		return err
	}

	if err := createCollections(conf, ctx); err != nil {
		return err
	}
	if err := createIdentities(conf, ctx); err != nil {
		return err
	}

	if err := initGlobalPlugins(ctx); err != nil {
		return err
	}
	if err := initAppPlugins(ctx); err != nil {
		return err
	}

	if err := initCollections(ctx); err != nil {
		return err
	}

	return nil
}

func createGlobalPlugins(conf *configs.Project, ctx *ProjectCtx) error {
	if err := createPwHashers(conf, ctx); err != nil {
		return err
	}
	if err := createSenders(conf, ctx); err != nil {
		return err
	}
	if err := createCryptoKeys(conf, ctx); err != nil {
		return err
	}
	if err := createStorages(conf, ctx); err != nil {
		return err
	}

	return nil
}

func createAppPlugins(conf *configs.Project, ctx *ProjectCtx) error {
	for n := range ctx.Apps {
		appCtx := ctx.Apps[n]
		appConf := conf.Apps[n]

		authenticators, err := createAuthenticators(&appConf)
		if err != nil {
			return err
		}

		authorizers, err := createAuthorizers(&appConf)
		if err != nil {
			return err
		}

		appCtx.Authenticators = authenticators
		appCtx.Authorizers = authorizers
	}

	return nil
}

func createStorages(conf *configs.Project, ctx *ProjectCtx) error {
	ctx.Storages = make(map[string]types.Storage)

	for i := range conf.StorageConfs {
		storageConf := conf.StorageConfs[i]
		connSess, err := storage.New(&storageConf)
		if err != nil {
			return fmt.Errorf("open connection session to storage '%s': %v", storageConf.Name, err)
		}

		ctx.Storages[storageConf.Name] = connSess
	}

	cleanupStorages(conf, ctx)
	return nil
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

func createCollections(conf *configs.Project, ctx *ProjectCtx) error {
	ctx.Collections = make(map[string]*collections.Collection)

	for _, collConf := range conf.CollConfs {
		coll, err := collections.Create(&collConf)
		if err != nil {
			return err
		}
		ctx.Collections[collConf.Name] = coll
	}

	return nil
}

func createPwHashers(conf *configs.Project, ctx *ProjectCtx) error {
	ctx.Hashers = make(map[string]pwhasherTypes.PwHasher)

	for i := range conf.HasherConfs {
		hasherConf := conf.HasherConfs[i]
		h, err := pwhasher.New(&conf.HasherConfs[i])
		if err != nil {
			return fmt.Errorf("cannot init hasher '%s': %v", hasherConf.Name, err)
		}

		ctx.Hashers[hasherConf.Name] = h
	}

	return nil
}

func createSenders(conf *configs.Project, ctx *ProjectCtx) error {
	ctx.Senders = make(map[string]senderTypes.Sender)

	for i := range conf.Senders {
		senderConf := conf.Senders[i]
		s, err := sender.New(&senderConf)
		if err != nil {
			return fmt.Errorf("cannot init sender '%s': %v", senderConf.Name, err)
		}

		ctx.Senders[senderConf.Name] = s
	}

	return nil
}

func createCryptoKeys(conf *configs.Project, ctx *ProjectCtx) error {
	ctx.CryptoKeys = make(map[string]cryptoKeyTypes.CryptoKey)

	for i := range conf.CryptoKeys {
		ckeyConf := conf.CryptoKeys[i]
		ckey, err := cryptokey.New(&ckeyConf)
		if err != nil {
			return fmt.Errorf("cannot init crypto key '%s': %v", ckeyConf.Name, err)
		}

		ctx.CryptoKeys[ckeyConf.Name] = ckey
	}

	return nil
}

func createApps(conf *configs.Project, ctx *ProjectCtx) error {
	ctx.Apps = make(map[string]*app.App, len(conf.Apps))

	for n := range conf.Apps {
		appConf := conf.Apps[n]

		ctx.Apps[n] = &app.App{
			PathPrefix: appConf.PathPrefix,
		}
	}

	return nil
}

func createIdentities(conf *configs.Project, ctx *ProjectCtx) error {
	for n := range ctx.Apps {
		appCtx := ctx.Apps[n]
		appConf := conf.Apps[n]
		i, err := identity.Create(&appConf.Identity, ctx.Collections)
		if err != nil {
			return err
		}
		appCtx.Identity = i
	}

	return nil
}

func createAuthenticators(app *configs.App) ([]authnTypes.Authenticator, error) {
	authenticators := make([]authnTypes.Authenticator, len(app.Authn))

	for i := range app.Authn {
		authnConf := app.Authn[i]
		authenticator, err := authn.New(&authnConf)
		if err != nil {
			return nil, err
		}

		authenticators[i] = authenticator
	}

	return authenticators, nil
}

func createAuthorizers(app *configs.App) (map[string]authzTypes.Authorizer, error) {
	authorizers := make(map[string]authzTypes.Authorizer, len(app.Authz))

	for i := range app.Authz {
		authzConf := app.Authz[i]
		authorizer, err := authz.New(&authzConf)
		if err != nil {
			return nil, err
		}

		authorizers[authzConf.Name] = authorizer
	}

	return authorizers, nil
}

func initCollections(ctx *ProjectCtx) error {
	for collName := range ctx.Collections {
		coll := ctx.Collections[collName]
		err := coll.Init(ctx.Collections)
		if err != nil {
			return err
		}
	}

	return nil
}

func initStorages(ctx *ProjectCtx) error {
	for _, s := range ctx.Storages {
		if err := s.Init(); err != nil {
			return err
		}
		return s.Ping()
	}

	return nil
}

func initPwHashers(ctx *ProjectCtx) error {
	for _, h := range ctx.Hashers {
		if err := h.Init(); err != nil {
			return err
		}
	}

	return nil
}

func initSenders(ctx *ProjectCtx) error {
	for _, s := range ctx.Senders {
		if err := s.Init(); err != nil {
			return err
		}
	}

	return nil
}

func initCryptoKeys(ctx *ProjectCtx) error {
	for _, k := range ctx.CryptoKeys {
		if err := k.Init(); err != nil {
			return err
		}
	}

	return nil
}

func initGlobalPlugins(ctx *ProjectCtx) error {
	if err := initStorages(ctx); err != nil {
		return err
	}
	if err := initPwHashers(ctx); err != nil {
		return err
	}
	if err := initSenders(ctx); err != nil {
		return err
	}
	if err := initCryptoKeys(ctx); err != nil {
		return err
	}

	return nil
}

func initAppPlugins(ctx *ProjectCtx) error {
	for appName, a := range ctx.Apps {
		if err := initAuthenticators(appName, a); err != nil {
			return err
		}
		if err := initAuthorizers(appName, a); err != nil {
			return err
		}
	}

	return nil
}

func initAuthenticators(appName string, app *app.App) error {
	for _, authenticator := range app.Authenticators {
		if err := authenticator.Init(appName); err != nil {
			return err
		}
	}

	return nil
}

func initAuthorizers(appName string, app *app.App) error {
	for _, authorizer := range app.Authorizers {
		if err := authorizer.Init(appName); err != nil {
			return err
		}
	}

	return nil
}
