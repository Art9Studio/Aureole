package pwbased

import (
	"github.com/mitchellh/mapstructure"
	authnTypes "gouth/adapters/authn/types"
	"gouth/adapters/pwhasher"
	"gouth/adapters/storage"
	"gouth/config"
	contextTypes "gouth/context/types"
)

type Config struct {
	MainHasher    string   `mapstructure:"main_hasher"`
	CompatHashers []string `mapstructure:"compat_hashers"`
	Collection    string   `mapstructure:"collection"`
	Identity      string   `mapstructure:"identity"`
	Password      string   `mapstructure:"password"`
}

type Ctx struct {
	ProjectContext *contextTypes.ProjectCtx
	PathPrefix     string
	PwHasher       pwhasher.PwHasher
	Storage        storage.ConnSession
	IdentityColl   storage.IdentityCollConfig
	Identity       string
	Password       string
}

func (p pwBasedAdapter) GetAuthnController(pathPrefix string, confMap *config.RawConfig, projectCtx *contextTypes.ProjectCtx) (authnTypes.Controller, error) {
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
