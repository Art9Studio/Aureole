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
	// RawProjectConfig represents raw settings for whole project. Based on raw
	// project settings initializes pure project settings, which describes below
	RawProjectConfig struct {
		APIVersion string                  `config:"api_version"`
		Apps       map[string]RawAppConfig `config:"apps"`
	}

	// RawAppConfig represents raw settings for one application. Based on raw
	// app settings initializes pure app settings, which describes below
	RawAppConfig struct {
		PathPrefix       string                              `config:"path_prefix"`
		RawStorageConfs  map[string]storage.RawStorageConfig `config:"storages"`
		Main             MainConfig                          `config:"main"`
		HashConf         RawHashConfig                       `config:"hasher"`
		StorageByFeature map[string]storage.ConnSession
		Hash             pwhash.PwHasher
	}

	// RawHashConfig represents raw settings for initializing hashers
	RawHashConfig struct {
		AlgName     string               `config:"alg"`
		RawHashConf pwhash.RawHashConfig `config:"settings"`
	}
)

type (
	// ProjectConfig represents settings for whole project
	ProjectConfig struct {
		APIVersion string
		Apps       map[string]AppConfig
	}

	// AppConfig represents settings for one application
	AppConfig struct {
		PathPrefix       string
		Main             MainConfig
		StorageByFeature map[string]storage.ConnSession
		Hash             pwhash.PwHasher
	}

	// MainConfig represents settings for authentication
	MainConfig struct {
		UseExistColl bool                    `config:"use_existent_collection"`
		UserColl     *storage.UserCollConfig `config:"user_collection"`
		AuthN        AuthNConfig             `config:"authN"`
		AuthZ        AuthZConfig             `config:"authZ"`
		Register     RegisterConfig          `config:"register"`
	}

	// AuthNConfig represents settings for authentication methods
	AuthNConfig struct {
		PasswdBased PasswordBasedConfig `config:"password_based"`
	}

	// AuthZConfig represents settings for authorization methods
	AuthZConfig struct {
		CookieConf CookieAuthConfig `config:"cookie"`
		Jwt        JWTConfig        `config:"jwt"`
	}

	// RegisterConfig represents settings for registering account
	RegisterConfig struct {
		LoginAfter bool              `config:"login_after"`
		AuthType   string            `config:"auth_type"`
		Fields     map[string]string `config:"fields"`
	}

	PasswordBasedConfig struct {
		UserUnique  string `config:"user_unique"`
		UserConfirm string `config:"user_confirm"`
	}

	CookieAuthConfig struct {
		StorageName string `config:"storage"`
		Domain      string `config:"domain"`
		Path        string `config:"path"`
		MaxAge      int    `config:"max_age"`
		IsSecure    bool   `config:"secure"`
		IsHttpOnly  bool   `config:"http_only"`
	}

	JWTConfig struct {
		Alg     string                   `config:"alg"`
		Keys    []map[string]interface{} `config:"keys"`
		KidAlg  string                   `config:"kid_alg"`
		Payload map[string]string        `config:"payload"`
	}
)

// Init loads settings for whole project into global object conf
func (c *ProjectConfig) Init() error {
	confLoader, err := configuro.NewConfig(
		configuro.WithLoadFromConfigFile("./config.yaml", true),
	)
	if err != nil {
		return fmt.Errorf("project config init: %v", err)
	}

	rawConf := &RawProjectConfig{}

	if err = confLoader.Load(rawConf); err != nil {
		return fmt.Errorf("project config init: %v", err)
	}

	for i := range rawConf.Apps {
		if app, ok := rawConf.Apps[i]; ok {
			if err = app.init(); err != nil {
				return err
			}
			rawConf.Apps[i] = app
		} else {
			return fmt.Errorf("project config init: cannot init app %s", i)
		}
	}

	if err = copier.Copy(c, rawConf); err != nil {
		return fmt.Errorf("project config init: %v", err)
	}

	return nil
}

// init initializes app
func (a *RawAppConfig) init() error {
	if err := a.initStorages(); err != nil {
		return fmt.Errorf("app init: %v", err)
	}

	if err := a.initUserColl(); err != nil {
		return fmt.Errorf("app init: %v", err)
	}

	if err := a.initPwHasher(); err != nil {
		return fmt.Errorf("app init: %v", err)
	}

	return nil
}

func (a *RawAppConfig) initStorages() error {
	storageFeatures := map[string][]string{}
	a.StorageByFeature = map[string]storage.ConnSession{}

	for storageName := range a.RawStorageConfs {
		if _, ok := storageFeatures[storageName]; !ok {
			storageFeatures[storageName] = []string{}
		}
	}

	usersStorage := a.Main.UserColl.StorageName
	storageFeatures[usersStorage] = append(storageFeatures[usersStorage], "users")

	if a.Main.AuthZ.CookieConf.StorageName != "" {
		sessionStorage := a.Main.AuthZ.CookieConf.StorageName
		storageFeatures[sessionStorage] = append(storageFeatures[sessionStorage], "sessions")
	}

	for storageName, features := range storageFeatures {
		connSess, err := storage.Open(a.RawStorageConfs[storageName], features)
		if err != nil {
			return fmt.Errorf("open connection session to storage '%s': %v", storageName, err)
		}

		for _, f := range features {
			a.StorageByFeature[f] = connSess
		}
	}

	return nil
}

func (a *RawAppConfig) initUserColl() error {
	usersStorage := a.StorageByFeature["users"]

	if a.Main.UserColl == nil {
		a.Main.UserColl = storage.NewUserCollConfig("users", "id", "username", "password")
	}
	isExists, err := usersStorage.IsCollExists(a.Main.UserColl.ToCollConfig())
	if err != nil {
		return err
	}

	if !a.Main.UseExistColl && !isExists {
		if err = usersStorage.CreateUserColl(*a.Main.UserColl); err != nil {
			return err
		}
	} else if a.Main.UseExistColl && !isExists {
		return errors.New("user collection is not found")
	}

	return nil
}

func (a *RawAppConfig) initPwHasher() error {
	h, err := pwhash.New(a.HashConf.AlgName, &a.HashConf.RawHashConf)
	if err != nil {
		return fmt.Errorf("cannot init hasher '%s': %v", a.HashConf.AlgName, err)
	}

	a.Hash = h
	return nil
}
