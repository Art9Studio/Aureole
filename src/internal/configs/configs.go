package configs

import (
	"fmt"
	"os"

	"aureole/pkg/configuro"

	"github.com/joho/godotenv"
)

type PluginConfig struct {
	Plugin string    `config:"plugin"`
	Name   string    `config:"name"`
	Config RawConfig `config:"config"`
}

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
		Internal       Internal       `config:"internal"`
		Auth           []PluginConfig `config:"auth"`
		Issuer         PluginConfig   `config:"issuer"`
		MFA            []PluginConfig `config:"mfa"`
		IDManager      PluginConfig   `config:"id_manager"`
		CryptoStorages []PluginConfig `config:"crypto_storages"`
		Storages       []PluginConfig `config:"storages"`
		CryptoKeys     []PluginConfig `config:"crypto_keys"`
		Senders        []PluginConfig `config:"senders"`
		RootPlugins    []PluginConfig `config:"root_plugins"`
	}

	Internal struct {
		SignKey string `config:"sign_key"`
		EncKey  string `config:"enc_key"`
		Storage string `config:"storage"`
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

	rawConf := &Project{}
	if err = confLoader.Load(rawConf); err != nil {
		return nil, fmt.Errorf("project config init: %v", err)
	}
	rawConf.setDefaults()

	return rawConf, nil
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
