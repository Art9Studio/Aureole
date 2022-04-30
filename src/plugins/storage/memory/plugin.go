package memory

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	_ "embed"
	"github.com/coocood/freecache"
	"github.com/mitchellh/mapstructure"

	"encoding/json"
	"errors"
)

//go:embed meta.yaml
var rawMeta []byte

var meta plugins.Meta

// init initializes package by register pluginCreator
func init() {
	meta = plugins.Repo.Register(rawMeta, pluginCreator{})
}

// pluginCreator represents plugin for bigcache storage
type pluginCreator struct {
}

type memory struct {
	pluginApi core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
	cache     *freecache.Cache
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

func (m memory) GetMetaData() plugins.Meta {
	return meta
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
