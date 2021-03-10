package config

import (
	"fmt"
	"github.com/sherifabdlnaby/configuro"
)

type (
	RawConfig = map[string]interface{}

	projectConfig struct {
		APIVersion  string        `с:"api_version"`
		Apps        []app         `с:"apps"`
		Storages    []storages    `с:"storages,omitempty"`
		Collections []collections `с:"collections,omitempty"`
		Hashers     []hashers     `с:"hashers,omitempty"`
		CryptoKeys  []cryptoKeys  `с:"crypto_keys,omitempty"`
	}

	app struct {
		PathPrefix    string          `с:"path_prefix"`
		AuthN         []authNConfig   `с:"authN"`
		AuthZ         authZ           `с:"authZ"`
		IdentityFlows []identityFlows `с:"identity_flows"`
	}

	authNConfig struct {
		Type   string    `с:"type"`
		Path   string    `с:"path,omitempty"`
		Config RawConfig `с:"config,omitempty"`
	}

	authZ struct {
		Type   string    `с:"type"`
		Config RawConfig `с:"config,omitempty"`
	}

	identityFlows struct {
		Type   string    `с:"type"`
		Config RawConfig `с:"config,omitempty"`
	}

	storages struct {
		Name   string    `с:"name"`
		Config RawConfig `с:"config,omitempty"`
	}

	collections struct {
		Type    string    `с:"type"`
		Name    string    `с:"name"`
		Storage string    `с:"storage"`
		Config  RawConfig `с:"config,omitempty"`
	}

	hashers struct {
		Type   string    `с:"type"`
		Name   string    `с:"name"`
		Config RawConfig `с:"config,omitempty"`
	}

	cryptoKeys struct {
		Type   string    `с:"type"`
		Driver string    `с:"driver"`
		Name   string    `с:"name"`
		Config RawConfig `с:"config,omitempty"`
	}
)

func LoadMainConfig(project *Project) error {
	confLoader, err := configuro.NewConfig(
		configuro.WithLoadFromConfigFile("./config.yaml", true),
		//configuro.KeyDelimiter(":"),
		configuro.Tag("c", "v"),
		//configuro.WithoutValidateByFunc(),
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
