package pwbased

import (
	"aureole/configs"
	"aureole/plugins/authn"
	"aureole/plugins/authn/types"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type (
	сonf struct {
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

func (c *сonf) setDefaults() {
	configs.SetDefault(&c.CompatHashers, []string{})
	c.Login.setDefaults()
	c.Register.setDefaults()
}

func (l *login) setDefaults() {
	configs.SetDefault(&l.Path, "/login")

	if l.FieldsMap == nil {
		l.FieldsMap = map[string]string{
			"identity": "{$.username}",
			"password": "{$.password}",
		}
	} else {
		if _, ok := l.FieldsMap["identity"]; !ok {
			l.FieldsMap["identity"] = "{$.username}"
		}

		if _, ok := l.FieldsMap["password"]; !ok {
			l.FieldsMap["password"] = "{$.password}"
		}
	}
}

func (r *register) setDefaults() {
	configs.SetDefault(&r.Path, "/register")

	if r.FieldsMap == nil {
		r.FieldsMap = map[string]string{
			"identity": "{$.username}",
			"password": "{$.password}",
		}
	} else {
		if _, ok := r.FieldsMap["identity"]; !ok {
			r.FieldsMap["identity"] = "{$.username}"
		}

		if _, ok := r.FieldsMap["password"]; !ok {
			r.FieldsMap["password"] = "{$.password}"
		}
	}
}

func (p pwBasedAdapter) Create(conf *configs.Authn) (types.Controller, error) {
	adapterConfMap := conf.Config
	adapterConf := &сonf{}

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

func initAdapter(conf *configs.Authn, adapterConf *сonf) (*pwBased, error) {
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
