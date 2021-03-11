package pwbased

import (
	"github.com/mitchellh/mapstructure"
	"gouth/authN"
	"gouth/config"
	"gouth/pwhasher"
	"gouth/storage"
)

type Config struct {
	MainHasher    string   `mapstructure:"main_hasher"`
	CompatHashers []string `mapstructure:"compat_hashers"`
	Collection    string   `mapstructure:"collection"`
	Identity      string   `mapstructure:"identity"`
	Password      string   `mapstructure:"password"`
}

type Ctx struct {
	ProjectContext *config.Project
	PathPrefix     string
	PwHasher       pwhasher.PwHasher
	Storage        storage.ConnSession
	IdentityColl   storage.IdentityCollConfig
	Identity       string
	Password       string
}

func (p pwBasedAdapter) GetAuthNController(pathPrefix string, confMap *config.RawConfig, projectCtx *config.Project) (authN.Controller, error) {
	controllerConfig := &Config{}
	err := mapstructure.Decode(confMap, controllerConfig)
	if err != nil {
		return nil, err
	}

	context := &Ctx{
		PathPrefix:     pathPrefix,
		ProjectContext: projectCtx,
		Identity:       controllerConfig.Identity,
		Password:       controllerConfig.Password,
	}
	context.PwHasher = projectCtx.Hashers[controllerConfig.MainHasher]
	context.IdentityColl = projectCtx.Collections[controllerConfig.Collection].(storage.IdentityCollConfig)
	context.Storage = projectCtx.Storages[context.IdentityColl.StorageName]

	return &pwBased{context}, nil
}
