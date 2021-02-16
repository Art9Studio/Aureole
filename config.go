package main

import (
	"errors"
	"gouth/hash"
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
	Session    storage.Session     `yaml:"-"`
	RawDB      storage.RawConnData `yaml:"storage"`
	Auth       AuthConfig          `yaml:"auth"`
	Hash       HashConfig          `yaml:"hash"`
}

// AuthConfig represents settings for authentication
type AuthConfig struct {
	UseExistentColl bool                    `yaml:"use_existent_collection"`
	UserColl        *storage.UserCollConfig `yaml:"user_collection"`
	Login           LoginConfig             `yaml:"login"`
	Register        RegisterConfig          `yaml:"register"`
}

// LoginConfig represents settings for logging into account
type LoginConfig struct {
	Payload map[string]string `yaml:"payload"`
	Fields  map[string]string `yaml:"fields"`
}

// RegisterConfig represents settings for registering account
type RegisterConfig struct {
	LoginAfter bool              `yaml:"login_after"`
	Fields     map[string]string `yaml:"fields"`
}

// HashConfig represents settings for hashing
type HashConfig struct {
	Algorithm string             `yaml:"algorithm"`
	RawHash   hash.RawHashConfig `yaml:"settings"`
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
	sess, err := storage.Open(a.RawDB)
	if err != nil {
		log.Panicf("app open session: %v", err)
	}

	a.Session = sess

	if err = a.initUserColl(); err != nil {
		log.Panicf("app init: %v", err)
	}
}

func (a *AppConfig) initUserColl() error {
	if a.Auth.UserColl == nil {
		a.Auth.UserColl = storage.NewUserCollConfig("users", "id", "username", "password")
	}
	isExists, err := a.Session.IsCollExists(a.Auth.UserColl.ToCollConfig())
	if err != nil {
		return err
	}

	if !a.Auth.UseExistentColl && !isExists {
		if err = a.Session.CreateUserColl(*a.Auth.UserColl); err != nil {
			return err
		}
	} else if a.Auth.UseExistentColl && !isExists {
		return errors.New("user collection is not found")
	}

	return nil
}
