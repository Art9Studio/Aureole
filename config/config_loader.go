package config

import (
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/sherifabdlnaby/configuro"
	"gouth/authN"
	"gouth/authN/types"
	"gouth/pwhasher"
	"gouth/storage"
)

type (
	RawConfig = map[string]interface{}

	projectConfig struct {
		APIVersion      string                         `config:"api_version"`
		Apps            []app                          `config:"apps"`
		StorageConfs    []storages                     `config:"storages,omitempty"`
		CollectionConfs []collections                  `config:"collections,omitempty"`
		HasherConfs     []hashers                      `config:"hashers,omitempty"`
		CryptoKeys      []cryptoKeys                   `config:"crypto_keys,omitempty"`
		Collections     map[string]interface{}         `config:"-"`
		Storages        map[string]storage.ConnSession `config:"-"`
		Hashers         map[string]pwhasher.PwHasher   `config:"-"`
	}

	app struct {
		PathPrefix       string             `config:"path_prefix"`
		AuthN            []authNConfig      `config:"authN"`
		AuthZ            authZ              `config:"authZ"`
		IdentityFlows    []identityFlows    `config:"identity_flows"`
		AuthNControllers []authN.Controller `config:"-"`
	}

	authNConfig struct {
		TypeName   string     `config:"type"`
		PathPrefix string     `config:"path_prefix,omitempty"`
		Config     RawConfig  `config:"config,omitempty"`
		Type       types.Type `config:"-"`
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

	if err = rawConf.init(); err != nil {
		return fmt.Errorf("project config init: %v", err)
	}
	if err = copier.Copy(project, rawConf); err != nil {
		return fmt.Errorf("project config init: %v", err)
	}

	return nil
}

func (conf *projectConfig) init() error {
	if err := conf.initStorages(); err != nil {
		return err
	}
	if err := conf.initCollections(); err != nil {
		return err
	}
	if err := conf.initPwHashers(); err != nil {
		return err
	}

	return nil
}

func (conf *projectConfig) initStorages() error {
	storageFeatures := map[string][]string{}
	storageConfMap := map[string]RawConfig{}

	for _, storageConf := range conf.StorageConfs {
		if _, ok := storageFeatures[storageConf.Name]; !ok {
			storageFeatures[storageConf.Name] = []string{}
			storageConfMap[storageConf.Name] = storageConf.Config
		}
	}

	for _, collConf := range conf.CollectionConfs {
		// todo: remove duplicate feature
		storageFeatures[collConf.Storage] = append(storageFeatures[collConf.Storage], collConf.Type)
	}

	for storageName, features := range storageFeatures {
		connSess, err := storage.Open(storageConfMap[storageName], features)
		if err != nil {
			return fmt.Errorf("open connection session to storage '%s': %v", storageName, err)
		}

		conf.Storages[storageName] = connSess
	}

	return nil
}

func old() {
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

func (conf *projectConfig) initCollections() error {
	for _, collConf := range conf.CollectionConfs {
		switch collConf.Type {
		case "identity":
			identityColl := storage.NewIdentityCollection(collConf.Storage, collConf.Config)
			identityStorage := conf.Storages[collConf.Storage]

			isExists, err := identityStorage.IsCollExists(identityColl.ToCollConfig())
			if err != nil {
				return err
			}

			useExistent := collConf.Config["use_existent"].(bool)
			if !useExistent && !isExists {
				if err = identityStorage.CreateUserColl(*identityColl); err != nil {
					return err
				}
			} else if useExistent && !isExists {
				return fmt.Errorf("identity collection '%s' is not found", collConf.Name)
			}

			conf.Collections[collConf.Name] = identityColl
		case "session":
			sessionColl := storage.NewSessionCollection(collConf.Storage, collConf.Config)
			sessionStorage := conf.Storages[collConf.Storage]

			isExists, err := sessionStorage.IsCollExists(sessionColl.ToCollConfig())
			if err != nil {
				return err
			}

			useExistent := collConf.Config["use_existent"].(bool)
			if !useExistent && !isExists {
				if err = sessionStorage.CreateUserColl(*sessionColl); err != nil {
					return err
				}
			} else if useExistent && !isExists {
				return fmt.Errorf("session collection '%s' is not found", collConf.Name)
			}

			conf.Collections[collConf.Name] = sessionColl
		}
	}

	return nil
}

func (conf *projectConfig) initPwHashers() error {
	for _, hasherConf := range conf.HasherConfs {
		h, err := pwhasher.New(hasherConf.Type, &hasherConf.Config)
		if err != nil {
			return fmt.Errorf("cannot init hasher '%s': %v", hasherConf.Name, err)
		}

		conf.Hashers[hasherConf.Name] = h
	}

	return nil
}
