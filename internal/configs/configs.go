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
		APIVersion string `config:"api_version"`
		PingPath   string `config:"-"`
		TestRun    bool   `config:"test_run"`
		Apps       []App  `config:"apps"`
	}

	App struct {
		Name           string         `config:"name"`
		Host           string         `config:"host"`
		PathPrefix     string         `config:"path_prefix"`
		AuthSessionExp int            `config:"auth_session_exp"`
		Service        Service        `config:"service"`
		Authn          []Authn        `config:"authN"`
		Authz          Authz          `config:"authZ"`
		SecondFactors  []SecondFactor `config:"2fa"`
		IDManager      IDManager      `config:"id_manager"`
		KeyStorages    []KeyStorage   `config:"key_storages"`
		Storages       []Storage      `config:"storages"`
		HasherConfs    []PwHasher     `config:"hashers"`
		CryptoKeys     []CryptoKey    `config:"crypto_keys"`
		Senders        []Sender       `config:"senders"`
		AdminConfs     []Admin        `config:"admin_plugins"`
	}

	Service struct {
		SignKey string `config:"sign_key"`
		EncKey  string `config:"enc_key"`
		Storage string `config:"storage"`
	}

	Authn struct {
		Type   string            `config:"type"`
		Filter map[string]string `config:"filter"`
		Config RawConfig         `config:"config"`
	}

	Authz struct {
		Type   string    `config:"type"`
		Config RawConfig `config:"config"`
	}

	SecondFactor struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}

	IDManager struct {
		Type   string    `config:"type"`
		Config RawConfig `config:"config"`
	}

	KeyStorage struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}

	Storage struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
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
		configuro.WithExpandEnvVars(),
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
