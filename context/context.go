package context

import (
	"fmt"
	"gouth/adapters/authn"
	authnTypes "gouth/adapters/authn/types"
	"gouth/adapters/pwhasher"
	"gouth/adapters/storage"
	"gouth/config"
	"gouth/context/types"
)

func InitContext(conf *config.ProjectConfig, ctx *types.ProjectCtx) interface{} {
	ctx.APIVersion = conf.APIVersion

	if err := initApps(conf, ctx); err != nil {
		return err
	}

	if err := initStorages(conf, ctx); err != nil {
		return err
	}
	//if err := conf.initCollections(); err != nil {
	//	return err
	//}
	if err := initPwHashers(conf, ctx); err != nil {
		return err
	}

	return nil
}

func initApps(conf *config.ProjectConfig, ctx *types.ProjectCtx) error {
	ctx.Apps = make([]types.App, len(conf.Apps))
	for i, app := range conf.Apps {
		authnControllers, err := initAuthnControllers(app.Authn, ctx)
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

func initAuthnControllers(authnList []config.AuthnConfig, ctx *types.ProjectCtx) ([]authnTypes.Controller, error) {
	res := make([]authnTypes.Controller, len(authnList))
	for i, authnItem := range authnList {
		controller, err := authn.New(&authnItem, ctx)
		if err != nil {
			return nil, err
		}
		res[i] = controller
	}
	return res, nil
}

func initStorages(conf *config.ProjectConfig, ctx *types.ProjectCtx) error {
	storageFeatures := map[string][]string{}
	storageConfMap := map[string]config.RawConfig{}

	for _, storageConf := range conf.StorageConfs {
		if _, ok := storageFeatures[storageConf.Name]; !ok {
			storageFeatures[storageConf.Name] = []string{}
			storageConfMap[storageConf.Name] = storageConf.Config
		}
	}

	for _, collConf := range conf.CollectionConfs {
		// todo: remove duplicate feature
		// todo: think about unused storages
		storageFeatures[collConf.Storage] = append(storageFeatures[collConf.Storage], collConf.Type)
	}

	ctx.Storages = make(map[string]storage.ConnSession)
	for storageName, features := range storageFeatures {
		connSess, err := storage.Open(storageConfMap[storageName], features)
		if err != nil {
			return fmt.Errorf("open connection session to storage '%s': %v", storageName, err)
		}

		ctx.Storages[storageName] = connSess
	}

	return nil
}

//func (conf *ProjectConfig) initCollections() error {
//	for _, collConf := range conf.CollectionConfs {
//		switch collConf.Type {
//		case "identity":
//			identityColl := storage.NewIdentityCollection(collConf.Storage, collConf.Config)
//			identityStorage := conf.Storages[collConf.Storage]
//
//			isExists, err := identityStorage.IsCollExists(identityColl.ToCollConfig())
//			if err != nil {
//				return err
//			}
//
//			useExistent := collConf.Config["use_existent"].(bool)
//			if !useExistent && !isExists {
//				if err = identityStorage.CreateUserColl(*identityColl); err != nil {
//					return err
//				}
//			} else if useExistent && !isExists {
//				return fmt.Errorf("identity collection '%s' is not found", collConf.Name)
//			}
//
//			conf.Collections[collConf.Name] = identityColl
//		case "session":
//			sessionColl := storage.NewSessionCollection(collConf.Storage, collConf.Config)
//			sessionStorage := conf.Storages[collConf.Storage]
//
//			isExists, err := sessionStorage.IsCollExists(sessionColl.ToCollConfig())
//			if err != nil {
//				return err
//			}
//
//			useExistent := collConf.Config["use_existent"].(bool)
//			if !useExistent && !isExists {
//				if err = sessionStorage.CreateUserColl(*sessionColl); err != nil {
//					return err
//				}
//			} else if useExistent && !isExists {
//				return fmt.Errorf("session collection '%s' is not found", collConf.Name)
//			}
//
//			conf.Collections[collConf.Name] = sessionColl
//		}
//	}
//
//	return nil
//}

func initPwHashers(conf *config.ProjectConfig, ctx *types.ProjectCtx) error {
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
