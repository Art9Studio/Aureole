package main

import (
	"gouth/storage"
	"log"

	"gopkg.in/yaml.v3"
)

// conf is global object that holds all project level settings variables

// ProjectConfig represents settings for whole project
type ProjectConfig struct {
	APIVersion string               `yaml:"api_version"`
	Apps       map[string]AppConfig `yaml:"apps"`
}

// AppConfig represents settings for one application
type AppConfig struct {
	PathPrefix string                    `yaml:"path_prefix"`
	Session    *storage.Session          `yaml:"-"`
	RawDB      storage.RawConnectionData `yaml:"storage"`
	Auth       AuthConfig                `yaml:"auth"`
}

//// DBConfig represents settings for database
//type DBConfig struct {
//	ConnURL  string           `yaml:"connection_url,omitempty"`
//	ConnConf adapters.ConnectionConfig `yaml:"connection_config,omitempty"`
//}

//// ConnectionConfig represents settings for set up connection with database
//type ConnectionConfig struct {
//	Driver   string            `yaml:"adapter"`
//	User     string            `yaml:"username"`
//	Password string            `yaml:"password"`
//	Host     string            `yaml:"host"`
//	Port     string            `yaml:"port"`
//	DBName   string            `yaml:"db_name"`
//	Opts     map[string]string `yaml:"options,omitempty"`
//}

// AuthConfig represents settings for authentication
type AuthConfig struct {
	UseCustomColl bool           `yaml:"use_custom_collection"`
	UserColl      UserCollConfig `yaml:"user_collection"`
	Login         LoginConfig    `yaml:"login"`
	Register      RegisterConfig `yaml:"register"`
}

// UserCollConfig represents settings for collection used to store users
type UserCollConfig struct {
	Name        string `yaml:"name"`
	PK          string `yaml:"pk,omitempty"`
	UserID      string `yaml:"user_id"`
	UserConfirm string `yaml:"user_confirm"`
}

// LoginConfig represents settings for logging into account
type LoginConfig struct {
	Fields  map[string]string `yaml:"fields"`
	Payload map[string]string `yaml:"payload"`
}

// RegisterConfig represents settings for registering account
type RegisterConfig struct {
	LoginAfter bool              `yaml:"login_after"`
	Fields     map[string]string `yaml:"fields"`
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
		}
	}
}

// init initializes app by creating table users
func (a *AppConfig) init() {
	sess, err := storage.Open(a.RawDB)
	if err != nil {
		log.Panicf("app open session: %v", err)
	}

	a.Session = &sess
}
