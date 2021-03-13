package pwbased

import (
	"github.com/mitchellh/mapstructure"
	authnTypes "gouth/adapters/authn/types"
	"gouth/adapters/pwhasher"
	"gouth/adapters/storage"
	"gouth/collections"
	"gouth/configs"
	contextTypes "gouth/context/types"
)

type Config struct {
	MainHasher    string   `mapstructure:"main_hasher"`
	CompatHashers []string `mapstructure:"compat_hashers"`
	Collection    string   `mapstructure:"collection"`
	Storage       string   `mapstructure:"storage"`
	Identity      string   `mapstructure:"identity"`
	Password      string   `mapstructure:"password"`
}

type Ctx struct {
	ProjectContext *contextTypes.ProjectCtx
	PathPrefix     string
	PwHasher       pwhasher.PwHasher
	Storage        storage.ConnSession
	IdentityColl   *collections.Collection
	Identity       string
	Password       string
}

func (p pwBasedAdapter) GetAuthnController(pathPrefix string, confMap *configs.RawConfig, projectCtx *contextTypes.ProjectCtx) (authnTypes.Controller, error) {
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
	context.IdentityColl = projectCtx.Collections[controllerConfig.Collection]
	context.Storage = projectCtx.Storages[controllerConfig.Storage]

	err = context.Storage.CheckFeaturesAvailable([]string{context.IdentityColl.Type})
	if err != nil {
		return nil, err
	}

	return &pwBased{context}, nil
}
