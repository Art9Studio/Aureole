package config

import (
	"fmt"
	"github.com/sherifabdlnaby/configuro"
)

type (
	RawConfig = map[string]interface{}

	ProjectConfig struct {
		APIVersion string `config:"api_version"`
		// todo: fix loading of apps. it is't array :(
		Apps            []app         `config:"apps"`
		StorageConfs    []storages    `config:"storages,omitempty"`
		CollectionConfs []collections `config:"collections,omitempty"`
		HasherConfs     []hashers     `config:"hashers,omitempty"`
		CryptoKeys      []cryptoKeys  `config:"crypto_keys,omitempty"`
		//Collections     map[string]interface{}         `config:"-"`
		//Storages        map[string]storage.ConnSession `config:"-"`
		//Hashers         map[string]pwhasher.PwHasher   `config:"-"`
	}

	app struct {
		PathPrefix    string          `config:"path_prefix"`
		Authn         []AuthnConfig   `config:"authn"`
		AuthZ         authZ           `config:"authZ"`
		IdentityFlows []identityFlows `config:"identity_flows"`
		//AuthnControllers []authnTypes.Controller `config:"-"`
	}

	AuthnConfig struct {
		TypeName   string    `config:"type"`
		PathPrefix string    `config:"path_prefix,omitempty"`
		Config     RawConfig `config:"config,omitempty"`
		//Type       types.Type `config:"-"`
	}

	authZ struct {
		Type   string    `config:"type"`
		Config RawConfig `config:"config,omitempty"`
	}

	identityFlows struct {
		Type   string    `config:"type"`
		Config RawConfig `config:"config,omitempty"`
	}

	storages struct {
		Name   string    `config:"name"`
		Config RawConfig `config:"config,omitempty"`
	}

	collections struct {
		Type    string    `config:"type"`
		Name    string    `config:"name"`
		Storage string    `config:"storage"`
		Config  RawConfig `config:"config,omitempty"`
	}

	hashers struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config,omitempty"`
	}

	cryptoKeys struct {
		Type   string    `config:"type"`
		Driver string    `config:"driver"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config,omitempty"`
	}
)

func LoadMainConfig() (*ProjectConfig, error) {
	confLoader, err := configuro.NewConfig(
		configuro.WithLoadFromConfigFile("./config.yaml", true),
		configuro.WithoutValidateByTags(),
	)
	if err != nil {
		return nil, fmt.Errorf("project config init: %v", err)
	}

	rawConf := &ProjectConfig{}
	if err = confLoader.Load(rawConf); err != nil {
		return nil, fmt.Errorf("project config init: %v", err)
	}
	if err = confLoader.Validate(rawConf); err != nil {
		return nil, fmt.Errorf("project config init: %v", err)
	}

	return rawConf, nil
}

//
//func old() {
//	isExists, err := usersStorage.IsCollExists(a.Main.UserColl.ToCollConfig())
//	if err != nil {
//		return err
//	}
//
//	if !a.Main.UseExistColl && !isExists {
//		if err = usersStorage.CreateUserColl(*a.Main.UserColl); err != nil {
//			return err
//		}
//	} else if a.Main.UseExistColl && !isExists {
//		return errors.New("user collection is not found")
//	}
//
//	return nil
//}
