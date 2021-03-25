package context

import (
	"aureole/configs"
	"aureole/context/types"
	"aureole/internal/collections"
	"aureole/internal/plugins/authn"
	authnTypes "aureole/internal/plugins/authn/types"
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
	err := initSenders(conf, ctx)

	return err
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
	ctx.Apps = make(map[string]types.App)

	for i, app := range conf.Apps {
		authnControllers, err := getAuthnControllers(app.Authn)
		if err != nil {
			return err
		}

		ctx.Apps[i] = types.App{
			PathPrefix:       app.PathPrefix,
			AuthnControllers: authnControllers,
		}
	}
	return nil
}

func getAuthnControllers(authnList []configs.Authn) ([]authnTypes.Controller, error) {
	controllers := make([]authnTypes.Controller, len(authnList))

	for i, authnItem := range authnList {
		controller, err := authn.New(&authnItem)
		if err != nil {
			return nil, err
		}

		controllers[i] = controller
	}

	return controllers, nil
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
