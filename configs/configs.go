package configs

import (
	"fmt"
	"github.com/sherifabdlnaby/configuro"
)

type Defaultable interface {
	setDefaults()
}

type (
	RawConfig = map[string]interface{}

	Project struct {
		APIVersion   string         `config:"api_version"`
		Apps         map[string]app `config:"apps"`
		StorageConfs []Storage      `config:"storages"`
		CollConfs    []Collection   `config:"collections"`
		HasherConfs  []PwHasher     `config:"hashers"`
		CryptoKeys   []cryptoKey    `config:"crypto_keys"`
	}

	app struct {
		PathPrefix string  `config:"path_prefix"`
		Authn      []Authn `config:"authN"`
		Authz      []Authz `config:"authZ"`
	}

	Authn struct {
		Type       string    `config:"type"`
		PathPrefix string    `config:"path_prefix"`
		AuthZ      string    `config:"authZ"`
		Config     RawConfig `config:"config"`
	}

	Authz struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}

	Storage struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}

	Collection struct {
		Type        string        `config:"type"`
		Name        string        `config:"name"`
		UseExistent bool          `config:"use_existent"`
		Spec        specification `config:"config"`
	}

	specification struct {
		Name      string            `config:"name"`
		Pk        string            `config:"pk"`
		FieldsMap map[string]string `config:"fields_map"`
	}

	PwHasher struct {
		Type   string    `config:"type"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}

	cryptoKey struct {
		Type   string    `config:"type"`
		Driver string    `config:"driver"`
		Name   string    `config:"name"`
		Config RawConfig `config:"config"`
	}
)

func (p *Project) setDefaults() {
	for i := range p.Apps {
		a := p.Apps[i]
		a.setDefaults()
		p.Apps[i] = a
	}
}

func (a *app) setDefaults() {
	for i := range a.Authn {
		a.Authn[i].setDefaults()
	}
}

func (authn *Authn) setDefaults() {
	if authn.PathPrefix == "" {
		authn.PathPrefix = "/"
	}
}

func (a *Authz) setDefaults() {
	panic("implement me")
}

func (s *Storage) setDefaults() {
	panic("implement me")
}

func (c *Collection) setDefaults() {
	panic("implement me")
}

func (s *specification) setDefaults() {
	panic("implement me")
}

func (h *PwHasher) setDefaults() {
	panic("implement me")
}

func (c *cryptoKey) setDefaults() {
	panic("implement me")
}

func LoadMainConfig() (*Project, error) {
	confLoader, err := configuro.NewConfig(
		configuro.WithLoadFromConfigFile("./config.yaml", true),
		configuro.WithoutValidateByTags(),
	)
	if err != nil {
		return nil, fmt.Errorf("project config init: %v", err)
	}

	rawConf := Project{}
	if err = confLoader.Load(&rawConf); err != nil {
		return nil, fmt.Errorf("project config init: %v", err)
	}

	rawConf.setDefaults()

	return &rawConf, nil
}
