package configs

import (
	"fmt"
	"os"

	"aureole/pkg/configuro"
	"github.com/joho/godotenv"
)

type (
	RawConfig = map[string]interface{}

	Project struct {
		APIVersion   string       `config:"api_version"`
		PingPath     string       `config:"ping_path"`
		TestRun      bool         `config:"test_run"`
		Apps         []App        `config:"apps"`
		StorageConfs []Storage    `config:"storages"`
		CollConfs    []Collection `config:"collections"`
		HasherConfs  []PwHasher   `config:"hashers"`
		CryptoKeys   []CryptoKey  `config:"crypto_keys"`
		Senders      []Sender     `config:"senders"`
		AdminConfs   []Admin      `config:"plugins"`
	}

	App struct {
		Name       string   `config:"name"`
		Host       string   `config:"host"`
		PathPrefix string   `config:"path_prefix"`
		Identity   Identity `config:"identity"`
		Authn      []Authn  `config:"authN"`
		Authz      []Authz  `config:"authZ"`
	}

	Identity struct {
		Collection string                `config:"collection"`
		Id         map[string]bool       `config:"id"`
		Username   map[string]bool       `config:"username"`
		Phone      map[string]bool       `config:"phone"`
		Email      map[string]bool       `config:"email"`
		Additional map[string]extraTrait `config:"additional"`
	}

	extraTrait struct {
		IsUnique   bool `config:"unique"`
		IsRequired bool `config:"required"`
		IsInternal bool `config:"internal"`
	}

	Authn struct {
		Type       string    `config:"type"`
		PathPrefix string    `config:"path_prefix"`
		AuthzName  string    `config:"authZ"`
		Config     RawConfig `config:"config"`
	}

	Authz struct {
		Type       string    `config:"type"`
		Name       string    `config:"name"`
		PathPrefix string    `config:"path_prefix"`
		Config     RawConfig `config:"config"`
	}

	Storage struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}

	Collection struct {
		Type   string   `config:"type"`
		Name   string   `config:"name"`
		Parent string   `config:"parent"`
		Spec   collSpec `config:"config"`
	}

	collSpec struct {
		Name      string                 `config:"name"`
		Pk        string                 `config:"pk"`
		FieldsMap map[string]interface{} `config:"fields_map"`
	}

	PwHasher struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}

	CryptoKey struct {
		Type       string    `config:"type"`
		Name       string    `config:"name"`
		PathPrefix string    `config:"path_prefix"`
		Config     RawConfig `config:"config"`
	}

	Sender struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}

	Admin struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}
)

func LoadMainConfig() (*Project, error) {
	var (
		confPath string
		ok       bool
	)

	_ = godotenv.Load("./.env")
	if confPath, ok = os.LookupEnv("AUREOLE_CONF_PATH"); !ok {
		confPath = "./config.yaml"
	}

	confLoader, err := configuro.NewConfig(
		configuro.WithLoadFromConfigFile(confPath, true),
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

/*
func validate() {
	// todo: at least 1 trait should be enabled, required, unique and credential
	// todo: trait can't be required if it internal
	// todo: trait can't be unique if not required
	// todo: check check if no non-existent keys (not in [enabled, unique, required, credential]) have been set in the config
	// todo: check bearer names [cookie, body, header, both]
	// todo: check curve, alg and kty in crypto keys generation
}
*/
