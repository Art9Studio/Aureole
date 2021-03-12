package context

import (
	"fmt"
	"gouth/adapters/authn"
	authnTypes "gouth/adapters/authn/types"
	"gouth/adapters/pwhasher"
	"gouth/adapters/storage"
	"gouth/collections"
	"gouth/configs"
	"gouth/context/types"
)

func InitContext(conf *configs.ProjectConfig, ctx *types.ProjectCtx) interface{} {
	ctx.APIVersion = conf.APIVersion

	if err := initApps(conf, ctx); err != nil {
		return err
	}

	if err := initStorages(conf, ctx); err != nil {
		return err
	}

	if err := initCollections(conf, ctx); err != nil {
		return err
	}

	if err := initPwHashers(conf, ctx); err != nil {
		return err
	}

	return nil
}

func initApps(conf *configs.ProjectConfig, ctx *types.ProjectCtx) error {
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

func getAuthnControllers(authnList []configs.AuthnConfig) ([]authnTypes.Controller, error) {
	res := make([]authnTypes.Controller, len(authnList))
	for i, authnItem := range authnList {
		controller, err := authn.New(&authnItem)
		if err != nil {
			return nil, err
		}
		res[i] = controller
	}
	return res, nil
}

func initStorages(conf *configs.ProjectConfig, ctx *types.ProjectCtx) error {
	ctx.Storages = make(map[string]storage.ConnSession)

	for _, storageConf := range conf.StorageConfs {
		connSess, err := storage.Open(storageConf.Config)
		if err != nil {
			return fmt.Errorf("open connection session to storage '%s': %v", storageConf.Name, err)
		}

		ctx.Storages[storageConf.Name] = connSess
	}

	// todo: think about unused storages
	return nil
}

func initCollections(conf *configs.ProjectConfig, ctx *types.ProjectCtx) error {
	ctx.Collections = make(map[string]collections.Collection)
	for _, collConf := range conf.CollectionConfs {
		coll, err := collections.NewCollection(collConf.Type, &collConf)
		if err != nil {
			return err
		}
		ctx.Collections[collConf.Name] = coll
	}

	return nil
}

func initPwHashers(conf *configs.ProjectConfig, ctx *types.ProjectCtx) error {
	ctx.Hashers = make(map[string]pwhasher.PwHasher)
	for _, hasherConf := range conf.HasherConfs {
		h, err := pwhasher.New(hasherConf.Type, &hasherConf.Config)
		if err != nil {
			return fmt.Errorf("cannot init hasher '%s': %v", hasherConf.Name, err)
		}

		ctx.Hashers[hasherConf.Name] = h
	}

	return nil
}
