package configs

import (
	"fmt"
	"github.com/sherifabdlnaby/configuro"
)

type (
	RawConfig = map[string]interface{}

	ProjectConfig struct {
		APIVersion      string         `config:"api_version"`
		Apps            map[string]App `config:"apps"`
		StorageConfs    []storages     `config:"storages,omitempty"`
		CollectionConfs []Collections  `config:"collections,omitempty"`
		HasherConfs     []hashers      `config:"hashers,omitempty"`
		CryptoKeys      []cryptoKeys   `config:"crypto_keys,omitempty"`
		//Collections     map[string]interface{}         `config:"-"`
		//Storages        map[string]storage.ConnSession `config:"-"`
		//Hashers         map[string]pwhasher.PwHasher   `config:"-"`
	}

	App struct {
		PathPrefix    string          `config:"path_prefix"`
		Authn         []AuthnConfig   `config:"authn"`
		AuthZ         authZ           `config:"authZ"`
		IdentityFlows []identityFlows `config:"identity_flows"`
		//AuthnControllers []authnTypes.Controller `config:"-"`
	}

	AuthnConfig struct {
		TypeName   string    `config:"type"`
		PathPrefix string    `config:"path_prefix,omitempty"`
		Config     RawConfig `config:"configs,omitempty"`
		//Type       types.Type `config:"-"`
	}

	authZ struct {
		Type   string    `config:"type"`
		Config RawConfig `config:"configs,omitempty"`
	}

	identityFlows struct {
		Type   string    `config:"type"`
		Config RawConfig `config:"configs,omitempty"`
	}

	storages struct {
		Name   string    `config:"name"`
		Config RawConfig `config:"configs,omitempty"`
	}

	Collections struct {
		Type        string `config:"type"`
		Name        string `config:"name"`
		UseExistent bool   `config:"use_existent"`
		Config      Config `config:"config"`
	}

	Config struct {
		Name      string            `config:"name"`
		Pk        string            `config:"pk"`
		FieldsMap map[string]string `config:"fields_map"`
	}

	hashers struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"configs,omitempty"`
	}

	cryptoKeys struct {
		Type   string    `config:"type"`
		Driver string    `config:"driver"`
		Name   string    `config:"name"`
		Config RawConfig `config:"configs,omitempty"`
	}
)

func LoadMainConfig() (*ProjectConfig, error) {
	confLoader, err := configuro.NewConfig(
		// todo: make pr which fix a "configFileErrIfNotFound" feature
		// In load.go:32 must be `if c.configFileErrIfNotFound && pathErr.Op == "open" {`
		// it has typo with excess `'` after "open"
		configuro.WithLoadFromConfigFile("./config.yaml", true),
		configuro.WithoutValidateByTags(),
	)
	if err != nil {
		return nil, fmt.Errorf("project configs init: %v", err)
	}

	rawConf := ProjectConfig{}
	if err = confLoader.Load(&rawConf); err != nil {
		return nil, fmt.Errorf("project configs init: %v", err)
	}
	if err = confLoader.Validate(rawConf); err != nil {
		return nil, fmt.Errorf("project configs init: %v", err)
	}

	return &rawConf, nil
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
