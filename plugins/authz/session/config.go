package session

import (
	"aureole/configs"
	"aureole/internal/plugins/authn"
	"aureole/internal/plugins/authz/types"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"time"
)

type config struct {
	Collection    string `mapstructure:"collection"`
	Storage       string `mapstructure:"storage"`
	Domain        string `mapstructure:"domain"`
	Path          string `mapstructure:"path"`
	MaxAge        int    `mapstructure:"max_age"`
	Secure        bool   `mapstructure:"secure"`
	HttpOnly      bool   `mapstructure:"http_only"`
	SameSite      string `mapstructure:"same_site"`
	CleanInterval int    `mapstructure:"clean_interval"`
}

func (s sessionAdapter) Create(conf *configs.Authz) (types.Authorizer, error) {
	adapterConfMap := conf.Config
	adapterConf := &config{}

	err := mapstructure.Decode(adapterConfMap, adapterConf)
	if err != nil {
		return nil, err
	}

	adapterConf.setDefaults()

	adapter, err := initAdapter(conf, adapterConf)
	if err != nil {
		return nil, err
	}

	err = adapter.Storage.CheckFeaturesAvailable([]string{adapter.Collection.Type})
	if err != nil {
		return nil, err
	}

	return adapter, nil
}

func initAdapter(conf *configs.Authz, adapterConf *config) (*session, error) {
	projectCtx := authn.Repository.ProjectCtx

	collection, ok := projectCtx.Collections[adapterConf.Collection]
	if !ok {
		return nil, fmt.Errorf("collection named '%s' is not declared", adapterConf.Collection)
	}

	storage, ok := projectCtx.Storages[adapterConf.Storage]
	if !ok {
		return nil, fmt.Errorf("storage named '%s' is not declared", adapterConf.Storage)
	}

	isCollExist, err := storage.IsCollExists(collection.Spec)
	if err != nil {
		return nil, err
	}

	if !isCollExist {
		err := storage.CreateSessionColl(collection.Spec)
		if err != nil {
			return nil, err
		}
	}
	storage.SetCleanInterval(time.Duration(adapterConf.CleanInterval) * time.Second)
	storage.StartCleaning(collection.Spec)

	return &session{
		Conf:           adapterConf,
		ProjectContext: projectCtx,
		Storage:        storage,
		Collection:     collection,
	}, nil
}
