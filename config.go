package main

import (
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/sherifabdlnaby/configuro"
	"gouth/pwhash"
	"gouth/storage"
)

type (
	RawConfig = map[string]interface{}

	RawProjectConfig struct {
		APIVersion  string                         `с:"api_version"`
		Apps        []App                          `с:"apps"`
		RawStorages []Storages                     `с:"storages,omitempty"`
		Collections []Collections                  `с:"collections,omitempty"`
		RawHashers  []Hashers                      `с:"hashers,omitempty"`
		CryptoKeys  []CryptoKeys                   `с:"crypto_keys,omitempty"`
		Storages    map[string]storage.ConnSession `c:"-"`
		Hashers     pwhash.PwHasher                `c:"-"`
	}

	ProjectConfig struct {
		APIVersion  string
		Apps        []App
		Collections []Collections
		Storages    map[string]storage.ConnSession
		Hashers     pwhash.PwHasher
	}

	App struct {
		PathPrefix    string          `с:"path_prefix"`
		AuthN         []AuthN         `с:"authN"`
		AuthZ         AuthZ           `с:"authZ"`
		IdentityFlows []IdentityFlows `с:"identity_flows"`
	}

	AuthN struct {
		Type   string    `с:"type"`
		URL    string    `с:"url,omitempty"`
		Config RawConfig `с:"config,omitempty"`
	}

	AuthZ struct {
		Type   string    `с:"type"`
		Config RawConfig `с:"config,omitempty"`
	}

	IdentityFlows struct {
		Type   string    `с:"type"`
		Config RawConfig `с:"config,omitempty"`
	}

	Storages struct {
		Name   string    `с:"name"`
		Config RawConfig `с:"config,omitempty"`
	}

	Collections struct {
		Type    string    `с:"type"`
		Name    string    `с:"name"`
		Storage string    `с:"storage"`
		Config  RawConfig `с:"config,omitempty"`
	}

	Hashers struct {
		Type   string    `с:"type"`
		Name   string    `с:"name"`
		Config RawConfig `с:"config,omitempty"`
	}

	CryptoKeys struct {
		Type   string    `с:"type"`
		Driver string    `с:"driver"`
		Name   string    `с:"name"`
		Config RawConfig `с:"config,omitempty"`
	}
)

func (c *ProjectConfig) Init() error {
	confLoader, err := configuro.NewConfig(
		configuro.WithLoadFromConfigFile("./config.yaml", true),
		configuro.KeyDelimiter(":"),
		configuro.Tag("c", "v"),
		configuro.WithoutValidateByFunc(),
		configuro.WithValidateByTags(),
	)
	if err != nil {
		return fmt.Errorf("project config init: %v", err)
	}

	rawConf := &RawProjectConfig{}
	if err = confLoader.Load(rawConf); err != nil {
		return fmt.Errorf("project config init: %v", err)
	}
	if err = confLoader.Validate(rawConf); err != nil {
		return fmt.Errorf("project config init: %v", err)
	}

	if err = rawConf.init(); err != nil {
		return err
	}
	if err = copier.Copy(c, rawConf); err != nil {
		return fmt.Errorf("project config init: %v", err)
	}

	return nil
}

func (rawC *RawProjectConfig) init() error {
	if err := rawC.initStorages(); err != nil {
		return fmt.Errorf("project init: %v", err)
	}
	if err := rawC.initCollections(); err != nil {
		return fmt.Errorf("project init: %v", err)
	}
	if err := rawC.initPwHashers(); err != nil {
		return fmt.Errorf("project init: %v", err)
	}

	return nil
}

func (rawC *RawProjectConfig) initStorages() error {
	return nil
}

func (rawC *RawProjectConfig) initCollections() error {
	return nil
}

func (rawC *RawProjectConfig) initPwHashers() error {
	return nil
}
