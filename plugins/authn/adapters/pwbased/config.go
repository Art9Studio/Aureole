package pwbased

import (
	"aureole/configs"
	contextTypes "aureole/context/types"
	"aureole/plugins/authn/types"
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

func (p pwBasedAdapter) Get(conf *configs.Authn, projectCtx *contextTypes.ProjectCtx) (types.Controller, error) {
	adapterConfMap := conf.Config
	adapterConf := &сonf{}
	err := mapstructure.Decode(adapterConfMap, adapterConf)
	if err != nil {
		return nil, err
	}

	adapter, err := initAdapter(conf, adapterConf, projectCtx)
	if err != nil {
		return nil, err
	}

	err = adapter.Storage.CheckFeaturesAvailable([]string{adapter.IdentityColl.Type})
	if err != nil {
		return nil, err
	}

	return adapter, nil
}

func initAdapter(conf *configs.Authn, adapterConf *сonf, projectCtx *contextTypes.ProjectCtx) (*pwBased, error) {
	return &pwBased{
		Conf:           adapterConf,
		PathPrefix:     conf.PathPrefix,
		ProjectContext: projectCtx,
		PwHasher:       projectCtx.Hashers[adapterConf.MainHasher],
		IdentityColl:   projectCtx.Collections[adapterConf.Collection],
		Storage:        projectCtx.Storages[adapterConf.Storage],
	}, nil
}
