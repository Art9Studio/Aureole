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
		APIVersion string                  `c:"api_version" v:"required,numeric"`
		Apps       map[string]RawAppConfig `c:"apps" v:"required,dive"`
	}

	// RawAppConfig represents raw settings for one application. Based on raw
	// app settings initializes pure app settings, which describes below
	RawAppConfig struct {
		PathPrefix       string                              `c:"path_prefix" v:"required"`
		RawStorageConfs  map[string]storage.RawStorageConfig `c:"storages" v:"required,dive,dive,keys,eq=connection_url|eq=connection_config,endkeys"`
		Main             MainConfig                          `c:"main" v:"required"`
		HashConf         RawHashConfig                       `c:"hasher" v:"required"`
		StorageByFeature map[string]storage.ConnSession
		Hash             pwhash.PwHasher
	}

	// RawHashConfig represents raw settings for initializing hashers
	RawHashConfig struct {
		AlgName     string               `c:"alg" v:"required,oneof=argon2 pbkdf2"`
		RawHashConf pwhash.RawHashConfig `c:"settings" v:"required"`
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
		Main             *MainConfig
		StorageByFeature map[string]storage.ConnSession
		Hash             pwhash.PwHasher
	}

	// MainConfig represents settings for authentication
	MainConfig struct {
		UseExistColl bool                    `c:"use_existent_collection"`
		UserColl     *storage.UserCollConfig `c:"user_collection" v:"required"`
		AuthN        *AuthNConfig            `c:"authN" v:"required"`
		AuthZ        *AuthZConfig            `c:"authZ" v:"required"`
		Register     *RegisterConfig         `c:"register" v:"required"`
	}

	// AuthNConfig represents settings for authentication methods
	AuthNConfig struct {
		PasswdBased *PasswordBasedConfig `c:"password_based" v:"required"`
	}

	// AuthZConfig represents settings for authorization methods
	AuthZConfig struct {
		CookieConf *CookieAuthConfig `c:"cookie" v:"required_without=Jwt"`
		Jwt        *JWTConfig        `c:"jwt" v:"required_without=CookieConf"`
	}

	// RegisterConfig represents settings for registering account
	RegisterConfig struct {
		LoginAfter bool              `c:"login_after" v:"required"`
		AuthType   string            `c:"auth_type" v:"required"`
		Fields     map[string]string `c:"fields" v:"required"`
	}

	PasswordBasedConfig struct {
		UserUnique  string `c:"user_unique" v:"required"`
		UserConfirm string `c:"user_confirm" v:"required"`
	}

	CookieAuthConfig struct {
		StorageName string `c:"storage" v:"required"`
		Domain      string `c:"domain" v:"required"`
		Path        string `c:"path" v:"required"`
		MaxAge      int    `c:"max_age" v:"required"`
		IsSecure    bool   `c:"secure"`
		IsHttpOnly  bool   `c:"http_only"`
	}

	JWTConfig struct {
		Alg     string                   `c:"alg" v:"required"`
		Keys    []map[string]interface{} `c:"keys" v:"required"`
		KidAlg  string                   `c:"kid_alg" v:"required"`
		Payload map[string]string        `c:"payload" v:"required"`
	}
)

// Init loads settings for whole project into global object conf
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

	if a.Main.AuthZ.CookieConf != nil {
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
