package main

import (
	"errors"
	"gouth/pwhash"
	"gouth/storage"
	"log"

	"gopkg.in/yaml.v3"
)

// ProjectConfig represents settings for whole project
type ProjectConfig struct {
	APIVersion string               `yaml:"api_version"`
	Apps       map[string]AppConfig `yaml:"apps"`
}

// AppConfig represents settings for one application
type AppConfig struct {
	PathPrefix   string                           `yaml:"path_prefix"`
	Conn         map[string]storage.ConnSession   `yaml:"-"`
	RawConnConfs map[string]storage.RawConnConfig `yaml:"storages"`
	Main         MainConfig                       `yaml:"main"`
	Hash         HashConfig                       `yaml:"hasher"`
}

// MainConfig represents settings for authentication
type MainConfig struct {
	UseExistColl bool                    `yaml:"use_existent_collection"`
	UserColl     *storage.UserCollConfig `yaml:"user_collection"`
	CookieConf   CookieAuthConfig        `yaml:"cookie_auth"`
	AuthN        AuthNConfig             `yaml:"authN"`
	AuthZ        AuthZConfig             `yaml:"authZ"`
	Register     RegisterConfig          `yaml:"register"`
}

type CookieAuthConfig struct {
	StorageName string `yaml:"storage"`
	Domain      string `yaml:"domain"`
	Path        string `yaml:"path"`
	MaxAge      int    `yaml:"max_age"`
	IsSecure    bool   `yaml:"secure"`
	IsHttpOnly  bool   `yaml:"http_only"`
}

// AuthNConfig represents settings for authentication methods
type AuthNConfig struct {
	PasswdBased PasswordBasedConfig `yaml:"password_based"`
}

// AuthZConfig represents settings for authorization methods
type AuthZConfig struct {
	Jwt JWTConfig `yaml:"jwt"`
}

// RegisterConfig represents settings for registering account
type RegisterConfig struct {
	LoginAfter bool              `yaml:"login_after"`
	AuthType   string            `yaml:"auth_type"`
	Fields     map[string]string `yaml:"fields"`
}

type PasswordBasedConfig struct {
	UserUnique  string `yaml:"user_unique"`
	UserConfirm string `yaml:"user_confirm"`
}

type JWTConfig struct {
	Alg     string                   `yaml:"alg"`
	Keys    []map[string]interface{} `yaml:"keys"`
	KidAlg  string                   `yaml:"kid_alg"`
	Payload map[string]string        `yaml:"payload"`
}

// HashConfig represents settings for hashing
type HashConfig struct {
	AlgName     string               `yaml:"alg"`
	RawHashConf pwhash.RawHashConfig `yaml:"settings"`
}

// Init loads settings for whole project into global object conf
func (c *ProjectConfig) Init(data []byte) {
	if err := yaml.Unmarshal(data, c); err != nil {
		log.Panicf("project config init: %v", err)
	}

	for i := range c.Apps {
		if app, ok := c.Apps[i]; ok {
			app.init()
			c.Apps[i] = app
		} else {
			log.Panicf("project config init: cannot init app %s", i)
		}
	}
}

// init initializes app by creating table users
func (a *AppConfig) init() {
	sess, err := storage.Open(a.RawConnConfs["Main DB"], []string{"users"})
	if err != nil {
		log.Panicf("app open session: %v", err)
	}

	a.Conn = make(map[string]storage.ConnSession)
	a.Conn["users"] = sess

	if err = a.initUserColl(); err != nil {
		log.Panicf("app init: %v", err)
	}
}

func (a *AppConfig) initUserColl() error {
	if a.Main.UserColl == nil {
		a.Main.UserColl = storage.NewUserCollConfig("users", "id", "username", "password")
	}
	isExists, err := a.Conn["users"].IsCollExists(a.Main.UserColl.ToCollConfig())
	if err != nil {
		return err
	}

	if !a.Main.UseExistColl && !isExists {
		if err = a.Conn["users"].CreateUserColl(*a.Main.UserColl); err != nil {
			return err
		}
	} else if a.Main.UseExistColl && !isExists {
		return errors.New("user collection is not found")
	}

	return nil
}
