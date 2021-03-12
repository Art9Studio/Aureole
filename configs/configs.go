package configs

import (
	"fmt"
	"github.com/sherifabdlnaby/configuro"
)

type (
	RawConfig = map[string]interface{}

	ProjectConfig struct {
		APIVersion   string         `config:"api_version"`
		Apps         map[string]app `config:"apps"`
		StorageConfs []storage      `config:"storages,omitempty"`
		CollConfs    []Collection   `config:"collections,omitempty"`
		HasherConfs  []hasher       `config:"hashers,omitempty"`
		CryptoKeys   []cryptoKey    `config:"crypto_keys,omitempty"`
	}

	app struct {
		PathPrefix    string          `config:"path_prefix"`
		Authn         []AuthnConfig   `config:"authn"`
		AuthZ         authZ           `config:"authZ"`
		IdentityFlows []identityFlows `config:"identity_flows"`
	}

	AuthnConfig struct {
		Type       string    `config:"type"`
		PathPrefix string    `config:"path_prefix,omitempty"`
		Config     RawConfig `config:"config,omitempty"`
	}

	authZ struct {
		Type   string    `config:"type"`
		Config RawConfig `config:"config,omitempty"`
	}

	identityFlows struct {
		Type   string    `config:"type"`
		Config RawConfig `config:"config,omitempty"`
	}

	storage struct {
		Name   string    `config:"name"`
		Config RawConfig `config:"config,omitempty"`
	}

	Collection struct {
		Type        string        `config:"type"`
		Name        string        `config:"name"`
		UseExistent bool          `config:"use_existent"`
		Spec        specification `config:"config"`
	}

	specification struct {
		Name      string            `config:"name"`
		Pk        string            `config:"pk"`
		FieldsMap map[string]string `config:"fields_map"`
	}

	hasher struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config,omitempty"`
	}

	cryptoKey struct {
		Type   string    `config:"type"`
		Driver string    `config:"driver"`
		Name   string    `config:"name"`
		Config RawConfig `config:"configs,omitempty"`
	}
)

func LoadMainConfig() (*ProjectConfig, error) {
	confLoader, err := configuro.NewConfig(
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
