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
	// common info
	PathPrefix string
	//
	PwHasher     pwhasher.PwHasher
	Storage      storage.ConnSession
	IdentityColl storage.IdentityCollConfig
	Identity     string
	Password     string
}

func (p pwBasedAdapter) GetAuthNController(pathPrefix string, configMap *config.RawConfig, projectContext *config.Project) (authN.Controller, error) {
	controllerConfig := Config{}

	err := mapstructure.Decode(configMap, &controllerConfig)

	// todo: init context by config. also use copier

	context := Ctx{}
	context.PathPrefix = pathPrefix
	context.ProjectContext = projectContext

	if err != nil {
		return nil, err
	}

	return &pwBased{&context}, nil
}
