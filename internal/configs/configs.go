package configs

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const (
	mimeYaml = "yaml"
	aureole  = "AUREOLE"
)

type PluginConfig struct {
	Plugin string    `mapstructure:"plugin" json:"plugin"`
	Name   string    `mapstructure:"name" json:"name"`
	Config RawConfig `mapstructure:"config" json:"config"`
}

type AuthPluginConfig struct {
	PluginConfig `mapstructure:",squash" json:",inline"`
	Filter       string `mapstructure:"filter" json:"filter"`
}

type (
	RawConfig = map[string]interface{}

	Project struct {
		APIVersion string `mapstructure:"api_version" json:"api_version"`
		PingPath   string `mapstructure:"-" json:"-"`
		TestRun    bool   `mapstructure:"test_run" json:"test_run"`
		Mode       string `mapstructure:"mode" json:"mode"`
		Apps       []App  `mapstructure:"apps" json:"apps"`
	}

	App struct {
		Name           string             `mapstructure:"name" json:"name"`
		Host           string             `mapstructure:"host" json:"host"`
		PathPrefix     string             `mapstructure:"path_prefix" json:"path_prefix"`
		AuthSessionExp int                `mapstructure:"auth_session_exp" json:"auth_session_exp"`
		Internal       Internal           `mapstructure:"internal" json:"internal"`
		Auth           []AuthPluginConfig `mapstructure:"auth" json:"auth"`
		Issuer         PluginConfig       `mapstructure:"issuer" json:"issuer"`
		MFA            []PluginConfig     `mapstructure:"mfa" json:"mfa"`
		IDManager      PluginConfig       `mapstructure:"id_manager" json:"id_manager"`
		CryptoStorages []PluginConfig     `mapstructure:"crypto_storages" json:"crypto_storages"`
		Storages       []PluginConfig     `mapstructure:"storages" json:"storages"`
		CryptoKeys     []PluginConfig     `mapstructure:"crypto_keys" json:"crypto_keys"`
		Senders        []PluginConfig     `mapstructure:"senders" json:"senders"`
		RootPlugins    []PluginConfig     `mapstructure:"root_plugins" json:"root_plugins"`
		ScratchCode    PluginConfig       `mapstructure:"scratch_code" json:"scratch_code"`
		AuthFilter     RawConfig          `mapstructure:"auth_filter" json:"auth_filter"`
	}

	Internal struct {
		SignKey string `mapstructure:"sign_key" json:"sign_key"`
		EncKey  string `mapstructure:"enc_key" json:"enc_key"`
		Storage string `mapstructure:"storage" json:"storage"`
	}
)

func LoadMainConfig() (*Project, error) {
	var (
		confPath string
	)
	_ = godotenv.Load("./.env")

	viper.AutomaticEnv()
	viper.SetEnvPrefix(aureole)

	if confPath = viper.GetString("conf_path"); confPath == "" {
		confPath = "/home/latala/aure/Aureole/config.light.yaml"
	}

	viper.SetConfigType(mimeYaml)
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
