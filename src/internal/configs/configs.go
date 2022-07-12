package configs

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type PluginConfig struct {
	Plugin string    `mapstructure:"plugin"`
	Name   string    `mapstructure:"name"`
	Config RawConfig `mapstructure:"config"`
}

type AuthPluginConfig struct {
	PluginConfig `mapstructure:",squash"`
	Filter       string `mapstructure:"filter"`
}

type (
	RawConfig = map[string]interface{}

	Project struct {
		APIVersion string `mapstructure:"api_version"`
		PingPath   string `mapstructure:"-"`
		TestRun    bool   `mapstructure:"test_run"`
		Apps       []App  `mapstructure:"apps"`
	}

	App struct {
		Name           string             `mapstructure:"name"`
		Host           string             `mapstructure:"host"`
		PathPrefix     string             `mapstructure:"path_prefix"`
		AuthSessionExp int                `mapstructure:"auth_session_exp"`
		Internal       Internal           `mapstructure:"internal"`
		Auth           []AuthPluginConfig `mapstructure:"auth"`
		Issuer         PluginConfig       `mapstructure:"issuer"`
		MFA            []PluginConfig     `mapstructure:"mfa"`
		IDManager      PluginConfig       `mapstructure:"id_manager"`
		CryptoStorages []PluginConfig     `mapstructure:"crypto_storages"`
		Storages       []PluginConfig     `mapstructure:"storages"`
		CryptoKeys     []PluginConfig     `mapstructure:"crypto_keys"`
		Senders        []PluginConfig     `mapstructure:"senders"`
		RootPlugins    []PluginConfig     `mapstructure:"root_plugins"`
	}

	Internal struct {
		SignKey string `mapstructure:"sign_key"`
		EncKey  string `mapstructure:"enc_key"`
		Storage string `mapstructure:"storage"`
	}
)

func LoadMainConfig() (*Project, error) {
	var (
		confPath string
	)
	_ = godotenv.Load("./.env")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("AUREOLE")

	if confPath = viper.GetString("conf_path"); confPath == "" {
		confPath = "./config.yaml"
	}

	viper.SetConfigType("yaml")
	viper.SetConfigFile(confPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("project config init: %v", err)
	}

	rawConf := &Project{}
	if err := viper.Unmarshal(&rawConf); err != nil {
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
