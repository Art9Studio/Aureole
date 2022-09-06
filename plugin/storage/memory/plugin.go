package memory

import (
	"aureole/configs"
	"aureole/internal/core"
	_ "embed"

	"github.com/coocood/freecache"
	"github.com/mitchellh/mapstructure"

	"encoding/json"
	"errors"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

// init initializes package by register pluginCreator
func init() {
	meta = core.StorageRepo.Register(rawMeta, Create)
}

type memory struct {
	pluginApi core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
	cache     *freecache.Cache
}

func Create(conf configs.PluginConfig) core.Storage {
	return &memory{rawConf: conf}
}

func (m *memory) Init(api core.PluginAPI) error {
	m.pluginApi = api
	PluginConf := &config{}
	if err := mapstructure.Decode(m.rawConf.Config, PluginConf); err != nil {
		return err
	}
	PluginConf.setDefaults()
	m.conf = PluginConf
	m.cache = freecache.NewCache(m.conf.Size * 1024 * 1024)
	return nil
}

func (m *memory) GetMetadata() core.Metadata {
	return meta
}

func (m *memory) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{}
}

func (m *memory) Set(k string, v interface{}, exp int) error {
	if k == "" || v == nil {
		return errors.New("memory key storage: key and value cannot be empty")
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return m.cache.Set([]byte(k), data, exp)
}

func (m *memory) Get(k string, v interface{}) (ok bool, err error) {
	if k == "" || v == nil {
		return false, errors.New("memory key storage: key and value cannot be empty")
	}

	data, err := m.cache.Get([]byte(k))
	if err != nil {
		if err == freecache.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return true, json.Unmarshal(data, v)
}

func (m *memory) Delete(k string) error {
	if k == "" {
		return errors.New("memory key storage: key and value cannot be empty")
	}
	m.cache.Del([]byte(k))
	return nil
}

func (m *memory) Exists(k string) (bool, error) {
	if k == "" {
		return false, errors.New("memory key storage: key and value cannot be empty")
	}
	return m.Get(k, new(interface{}))
}

func (m *memory) Close() error {
	m.cache.Clear()
	return nil
}
