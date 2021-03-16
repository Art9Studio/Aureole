package configs

import (
	"fmt"
	"github.com/sherifabdlnaby/configuro"
)

type (
	RawConfig = map[string]interface{}

	Project struct {
		APIVersion   string         `config:"api_version"`
		Apps         map[string]app `config:"apps"`
		StorageConfs []Storage      `config:"storages,omitempty"`
		CollConfs    []Collection   `config:"collections,omitempty"`
		HasherConfs  []Hasher       `config:"hashers,omitempty"`
		CryptoKeys   []cryptoKey    `config:"crypto_keys,omitempty"`
	}

	app struct {
		PathPrefix    string          `config:"path_prefix"`
		Authn         []Authn         `config:"authN"`
		Authz         []Authz         `config:"authZ"`
		IdentityFlows []identityFlows `config:"identity_flows"`
	}

	Authn struct {
		Type       string    `config:"type"`
		PathPrefix string    `config:"path_prefix,omitempty"`
		AuthZ      string    `config:"authZ,omitempty"`
		Config     RawConfig `config:"config,omitempty"`
	}

	Authz struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config,omitempty"`
	}

	identityFlows struct {
		Type       string    `config:"type"`
		PathPrefix string    `config:"path_prefix,omitempty"`
		Config     RawConfig `config:"config,omitempty"`
	}

	Storage struct {
		Type   string    `config:"type"`
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

	Hasher struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config,omitempty"`
	}

	cryptoKey struct {
		Type   string    `config:"type"`
		Driver string    `config:"driver"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config,omitempty"`
	}
)

func LoadMainConfig() (*Project, error) {
	confLoader, err := configuro.NewConfig(
		configuro.WithLoadFromConfigFile("./config.yaml", true),
		configuro.WithoutValidateByTags(),
	)
	if err != nil {
		return nil, fmt.Errorf("project config init: %v", err)
	}

	rawConf := Project{}
	if err = confLoader.Load(&rawConf); err != nil {
		return nil, fmt.Errorf("project config init: %v", err)
	}
	if err = confLoader.Validate(rawConf); err != nil {
		return nil, fmt.Errorf("project config init: %v", err)
	}

	return &rawConf, nil
}
