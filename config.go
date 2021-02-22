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
	PathPrefix string              `yaml:"path_prefix"`
	Session    storage.ConnSession     `yaml:"-"`
	Main       MainConfig         `yaml:"auth"`
	Hash       HashConfig          `yaml:"pwhash"`
	// Raw data
	RawConnConfig storage.RawConnConfig `yaml:"storage"`
}

// MainConfig represents settings for authentication
type MainConfig struct {
	UseExistentColl bool                    `yaml:"use_existent_collection"`
	UserColl        *storage.UserCollConfig `yaml:"user_collection"`
	AuthZ           AuthZConfig             `yaml:"authZ"`
	AuthN           AuthNConfig             `yaml:"authN"`
	Register        RegisterConfig          `yaml:"register"`
}

// AuthNConfig represents settings for authentication methods
type AuthNConfig struct {
	PasswordBased PasswordBasedConfig `yaml:"password_based"`
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
	Alg     string            `yaml:"alg"`
	Keys    []string          `yaml:"keys"`
	KidAlg  string            `yaml:"kid_alg"`
	Payload map[string]string `yaml:"payload"`
}

// HashConfig represents settings for hashing
type HashConfig struct {
	Algorithm string               `yaml:"algorithm"`
	RawHash   pwhash.RawHashConfig `yaml:"settings"`
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
	sess, err := storage.Open(a.RawConnConfig)
	if err != nil {
		log.Panicf("app open session: %v", err)
	}

	a.Session = sess

	if err = a.initUserColl(); err != nil {
		log.Panicf("app init: %v", err)
	}
}

func (a *AppConfig) initUserColl() error {
	if a.Main.UserColl == nil {
		a.Main.UserColl = storage.NewUserCollConfig("users", "id", "username", "password")
	}
	isExists, err := a.Session.IsCollExists(a.Main.UserColl.ToCollConfig())
	if err != nil {
		return err
	}

	if !a.Main.UseExistentColl && !isExists {
		if err = a.Session.CreateUserColl(*a.Main.UserColl); err != nil {
			return err
		}
	} else if a.Main.UseExistentColl && !isExists {
		return errors.New("user collection is not found")
	}

	return nil
}
