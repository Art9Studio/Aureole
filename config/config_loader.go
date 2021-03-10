package config

import (
	"fmt"
	"github.com/sherifabdlnaby/configuro"
)

type (
	RawConfig = map[string]interface{}

	projectConfig struct {
		APIVersion  string        `config:"api_version"`
		Apps        []app         `config:"apps"`
		Storages    []storages    `config:"storages,omitempty"`
		Collections []collections `config:"collections,omitempty"`
		Hashers     []hashers     `config:"hashers,omitempty"`
		CryptoKeys  []cryptoKeys  `config:"crypto_keys,omitempty"`
	}

	app struct {
		PathPrefix    string          `config:"path_prefix"`
		AuthN         []authNConfig   `config:"authN"`
		AuthZ         authZ           `config:"authZ"`
		IdentityFlows []identityFlows `config:"identity_flows"`
	}

	authNConfig struct {
		Type   string    `config:"type"`
		Path   string    `config:"path,omitempty"`
		Config RawConfig `config:"config,omitempty"`
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

func LoadMainConfig(project *Project) error {
	confLoader, err := configuro.NewConfig(
		configuro.WithLoadFromConfigFile("./config.yaml", true),
		configuro.WithoutValidateByTags(),
	)
	if err != nil {
		return fmt.Errorf("project config init: %v", err)
	}

	rawConf := &projectConfig{}
	if err = confLoader.Load(rawConf); err != nil {
		return fmt.Errorf("project config init: %v", err)
	}
	if err = confLoader.Validate(rawConf); err != nil {
		return fmt.Errorf("project config init: %v", err)
	}
	// todo: init project
	// project

	//if err = rawConf.init(); err != nil {
	//	return err
	//}
	//if err = copier.Copy(c, rawConf); err != nil {
	//	return fmt.Errorf("project config init: %v", err)
	//}

	return nil
}

//
//func (rawC *RawProjectConfig) init() error {
//	if err := rawC.initStorages(); err != nil {
//		return fmt.Errorf("project init: %v", err)
//	}
//	if err := rawC.initCollections(); err != nil {
//		return fmt.Errorf("project init: %v", err)
//	}
//	if err := rawC.initPwHashers(); err != nil {
//		return fmt.Errorf("project init: %v", err)
//	}
//
//	return nil
//}
//
//func (rawC *RawProjectConfig) initStorages() error {
//	return nil
//}
//
//func (rawC *RawProjectConfig) initCollections() error {
//	return nil
//}
//
//func (rawC *RawProjectConfig) initPwHashers() error {
//	return nil
//}
