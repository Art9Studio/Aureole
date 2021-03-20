package pwbased

import (
	"aureole/configs"
	"aureole/plugins/authn"
	"aureole/plugins/authn/types"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type (
	сonfig struct {
		MainHasher    string   `mapstructure:"main_hasher"`
		CompatHashers []string `mapstructure:"compat_hashers"`
		Collection    string   `mapstructure:"collection"`
		Storage       string   `mapstructure:"storage"`
		Login         login    `mapstructure:"login"`
		Register      register `mapstructure:"register"`
	}

	login struct {
		Path      string            `mapstructure:"path"`
		FieldsMap map[string]string `mapstructure:"fields_map"`
	}

	register struct {
		Path         string            `mapstructure:"path"`
		IsLoginAfter bool              `mapstructure:"login_after"`
		FieldsMap    map[string]string `mapstructure:"fields_map"`
	}
)

func (p pwBasedAdapter) Create(conf *configs.Authn) (types.Controller, error) {
	adapterConfMap := conf.Config
	adapterConf := &сonfig{}

	err := mapstructure.Decode(adapterConfMap, adapterConf)
	if err != nil {
		return nil, err
	}

	adapterConf.setDefaults()

	adapter, err := initAdapter(conf, adapterConf)
	if err != nil {
		return nil, err
	}

	err = adapter.Storage.CheckFeaturesAvailable([]string{adapter.IdentityColl.Type})
	if err != nil {
		return nil, err
	}

	return adapter, nil
}

func initAdapter(conf *configs.Authn, adapterConf *сonfig) (*pwBased, error) {
	projectCtx := authn.Repository.ProjectCtx

	hasher, ok := projectCtx.Hashers[adapterConf.MainHasher]
	if !ok {
		return nil, fmt.Errorf("hasher named '%s' is not declared", adapterConf.MainHasher)
	}

	collection, ok := projectCtx.Collections[adapterConf.Collection]
	if !ok {
		return nil, fmt.Errorf("collection named '%s' is not declared", adapterConf.Collection)
	}

	storage, ok := projectCtx.Storages[adapterConf.Storage]
	if !ok {
		return nil, fmt.Errorf("storage named '%s' is not declared", adapterConf.Storage)
	}

	return &pwBased{
		Conf:           adapterConf,
		PathPrefix:     conf.PathPrefix,
		ProjectContext: projectCtx,
		PwHasher:       hasher,
		IdentityColl:   collection,
		Storage:        storage,
	}, nil
}
