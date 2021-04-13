package configs

import (
	"fmt"
	"github.com/sherifabdlnaby/configuro"
)

type (
	RawConfig = map[string]interface{}

	Project struct {
		APIVersion   string         `config:"api_version"`
		Apps         map[string]App `config:"apps"`
		StorageConfs []Storage      `config:"storages"`
		CollConfs    []Collection   `config:"collections"`
		HasherConfs  []PwHasher     `config:"hashers"`
		CryptoKeys   []CryptoKey    `config:"crypto_keys"`
		Senders      []Sender       `config:"senders"`
	}

	App struct {
		PathPrefix string  `config:"path_prefix"`
		Authn      []Authn `config:"authN"`
		Authz      []Authz `config:"authZ"`
	}

	Authn struct {
		Type       string    `config:"type"`
		PathPrefix string    `config:"path_prefix"`
		AuthzName  string    `config:"authZ"`
		Config     RawConfig `config:"config"`
	}

	Authz struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}

	Storage struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
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

	PwHasher struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}

	CryptoKey struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}

	Sender struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}
)

func LoadMainConfig() (*Project, error) {
	confLoader, err := configuro.NewConfig(
		configuro.WithLoadFromConfigFile("./config.yaml", true),
		configuro.WithoutValidateByTags(),
		configuro.WithLoadDotEnv(".env"),
	)
	if err != nil {
		return nil, fmt.Errorf("project config init: %v", err)
	}

	rawConf := Project{}
	if err = confLoader.Load(&rawConf); err != nil {
		return nil, fmt.Errorf("project config init: %v", err)
	}

	rawConf.setDefaults()

	return &rawConf, nil
}
