package context

import (
	"aureole/configs"
	"aureole/context/types"
	"aureole/internal/collections"
	"aureole/internal/plugins/authn"
	authnTypes "aureole/internal/plugins/authn/types"
	"aureole/internal/plugins/authz"
	authzTypes "aureole/internal/plugins/authz/types"
	"aureole/internal/plugins/cryptokey"
	ckeysTypes "aureole/internal/plugins/cryptokey/types"
	"aureole/internal/plugins/pwhasher"
	pwhasherTypes "aureole/internal/plugins/pwhasher/types"
	"aureole/internal/plugins/sender"
	senderTypes "aureole/internal/plugins/sender/types"
	"aureole/internal/plugins/storage"
	storageTypes "aureole/internal/plugins/storage/types"
	"fmt"
)

func InitContext(conf *configs.Project, ctx *types.ProjectCtx) error {
	ctx.APIVersion = conf.APIVersion

	if err := createStorages(conf, ctx); err != nil {
		return err
	}
	if err := createCollections(conf, ctx); err != nil {
		return err
	}
	if err := createPwHashers(conf, ctx); err != nil {
		return err
	}
	if err := createSenders(conf, ctx); err != nil {
		return err
	}
	if err := createCryptoKeys(conf, ctx); err != nil {
		return err
	}
	if err := createApps(conf, ctx); err != nil {
		return err
	}

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

	return initApps(ctx)
}

func createStorages(conf *configs.Project, ctx *types.ProjectCtx) error {
	ctx.Storages = make(map[string]storageTypes.Storage)

	for _, storageConf := range conf.StorageConfs {
		connSess, err := storage.New(&storageConf)
		if err != nil {
			return fmt.Errorf("open connection session to storage '%s': %v", storageConf.Name, err)
		}

		ctx.Storages[storageConf.Name] = connSess
	}

	return nil
}

func createCollections(conf *configs.Project, ctx *types.ProjectCtx) error {
	ctx.Collections = make(map[string]*collections.Collection)

	for _, collConf := range conf.CollConfs {
		coll := collections.New(&collConf)
		ctx.Collections[collConf.Name] = coll
	}

	return nil
}

func createPwHashers(conf *configs.Project, ctx *types.ProjectCtx) error {
	ctx.Hashers = make(map[string]pwhasherTypes.PwHasher)

	for _, hasherConf := range conf.HasherConfs {
		h, err := pwhasher.New(&hasherConf)
		if err != nil {
			return fmt.Errorf("cannot init hasher '%s': %v", hasherConf.Name, err)
		}

		ctx.Hashers[hasherConf.Name] = h
	}

	return nil
}

func createSenders(conf *configs.Project, ctx *types.ProjectCtx) error {
	ctx.Senders = make(map[string]senderTypes.Sender)

	for _, senderConf := range conf.Senders {
		s, err := sender.New(&senderConf)
		if err != nil {
			return fmt.Errorf("cannot init sender '%s': %v", senderConf.Name, err)
		}

		ctx.Senders[senderConf.Name] = s
	}

	return nil
}

func createCryptoKeys(conf *configs.Project, ctx *types.ProjectCtx) error {
	ctx.CryptoKeys = make(map[string]ckeysTypes.CryptoKey)

	for _, ckeyConf := range conf.CryptoKeys {
		ckey, err := cryptokey.New(&ckeyConf)
		if err != nil {
			return fmt.Errorf("cannot init crypto key '%s': %v", ckeyConf.Name, err)
		}

		ctx.CryptoKeys[ckeyConf.Name] = ckey
	}

	return nil
}

func createApps(conf *configs.Project, ctx *types.ProjectCtx) error {
	ctx.Apps = make(map[string]*types.App)

	for appName, app := range conf.Apps {
		authenticators, err := createAuthenticators(&app, appName)
		if err != nil {
			return err
		}

		authorizers, err := createAuthorizers(&app)
		if err != nil {
			return err
		}

		ctx.Apps[appName] = &types.App{
			PathPrefix:     app.PathPrefix,
			Authorizers:    authorizers,
			Authenticators: authenticators,
		}
	}

	return nil
}

func createAuthenticators(app *configs.App, appName string) ([]authnTypes.Authenticator, error) {
	authenticators := make([]authnTypes.Authenticator, len(app.Authn))

	for i, authnItem := range app.Authn {
		authenticator, err := authn.New(appName, &authnItem)
		if err != nil {
			return nil, err
		}

		authenticators[i] = authenticator
	}

	return authenticators, nil
}

func createAuthorizers(app *configs.App) (map[string]authzTypes.Authorizer, error) {
	authorizers := make(map[string]authzTypes.Authorizer, len(app.Authz))

	for _, authzItem := range app.Authz {
		authorizer, err := authz.New(&authzItem)
		if err != nil {
			return nil, err
		}

		authorizers[authzItem.Name] = authorizer
	}

	return authorizers, nil
}

func initStorages(ctx *types.ProjectCtx) error {
	for _, storage := range ctx.Storages {
		if err := storage.Initialize(); err != nil {
			return err
		}
	}

	return nil
}

func initPwHashers(ctx *types.ProjectCtx) error {
	for _, hasher := range ctx.Hashers {
		if err := hasher.Initialize(); err != nil {
			return err
		}
	}

	return nil
}

func initSenders(ctx *types.ProjectCtx) error {
	for _, sender := range ctx.Senders {
		if err := sender.Initialize(); err != nil {
			return err
		}
	}

	return nil
}

func initCryptoKeys(ctx *types.ProjectCtx) error {
	for _, cryptoKey := range ctx.CryptoKeys {
		if err := cryptoKey.Initialize(); err != nil {
			return err
		}
	}

	return nil
}

func initApps(ctx *types.ProjectCtx) error {
	for _, app := range ctx.Apps {
		if err := initAuthenticators(app); err != nil {
			return err
		}
		return initAuthorizers(app)
	}

	return nil
}

func initAuthenticators(app *types.App) error {
	for _, authenticator := range app.Authenticators {
		if err := authenticator.Initialize(); err != nil {
			return err
		}
	}

	return nil
}

func initAuthorizers(app *types.App) error {
	for _, authorizer := range app.Authorizers {
		if err := authorizer.Initialize(); err != nil {
			return err
		}
	}

	return nil
}
