package context

import (
	"aureole/configs"
	"aureole/context/types"
	"aureole/internal/collections"
	"aureole/internal/plugins/authn"
	authnTypes "aureole/internal/plugins/authn/types"
	"aureole/internal/plugins/authz"
	authzTypes "aureole/internal/plugins/authz/types"
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

	if err := initStorages(conf, ctx); err != nil {
		return err
	}
	if err := initCollections(conf, ctx); err != nil {
		return err
	}
	if err := initPwHashers(conf, ctx); err != nil {
		return err
	}
	if err := initApps(conf, ctx); err != nil {
		return err
	}

	return initSenders(conf, ctx)
}

func initStorages(conf *configs.Project, ctx *types.ProjectCtx) error {
	ctx.Storages = make(map[string]storageTypes.Storage)

	for _, storageConf := range conf.StorageConfs {
		connSess, err := storage.New(&storageConf)
		if err != nil {
			return fmt.Errorf("open connection session to storage '%s': %v", storageConf.Name, err)
		}

		err = connSess.Ping()
		if err != nil {
			return fmt.Errorf("trying to ping storage '%s' was failed: %v", storageConf.Name, err)
		}

		ctx.Storages[storageConf.Name] = connSess
	}

	cleanupConnections(conf, ctx)
	return nil
}

func cleanupConnections(conf *configs.Project, ctx *types.ProjectCtx) {
	isUsedStorage := make(map[string]bool)

	for storageName := range ctx.Storages {
		isUsedStorage[storageName] = false

		for _, app := range conf.Apps {
			for _, authzItem := range app.Authz {
				if storageName == authzItem.Config["storage"] {
					isUsedStorage[storageName] = true
					break
				}
			}

			for _, authnItem := range app.Authn {
				if storageName == authnItem.Config["storage"] {
					isUsedStorage[storageName] = true
					break
				}
			}
		}
	}

	for storageName, isUse := range isUsedStorage {
		if !isUse {
			delete(ctx.Storages, storageName)
		}
	}
}

func initCollections(conf *configs.Project, ctx *types.ProjectCtx) error {
	ctx.Collections = make(map[string]*collections.Collection)

	for _, collConf := range conf.CollConfs {
		coll := collections.New(collConf.Type, &collConf)
		ctx.Collections[collConf.Name] = coll
	}

	return nil
}

func initPwHashers(conf *configs.Project, ctx *types.ProjectCtx) error {
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

func initApps(conf *configs.Project, ctx *types.ProjectCtx) error {
	ctx.Apps = make(map[string]*types.App)

	for appName, app := range conf.Apps {
		authorizers, err := getAuthorizers(&app)
		if err != nil {
			return err
		}

		ctx.Apps[appName] = &types.App{
			PathPrefix:  app.PathPrefix,
			Authorizers: authorizers,
		}

		authenticators, err := getAuthenticators(&app, appName)
		if err != nil {
			return err
		}

		ctx.Apps[appName].Authenticators = authenticators
	}
	return nil
}

func getAuthenticators(app *configs.App, appName string) ([]authnTypes.Authenticator, error) {
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

func getAuthorizers(app *configs.App) (map[string]authzTypes.Authorizer, error) {
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

func initSenders(conf *configs.Project, ctx *types.ProjectCtx) error {
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
