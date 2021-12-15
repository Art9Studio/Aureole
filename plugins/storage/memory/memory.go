package memory

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"encoding/json"
	"errors"

	"github.com/coocood/freecache"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "7662"

type memory struct {
	pluginApi core.PluginAPI
	rawConf   *configs.Storage
	conf      *config
	cache     *freecache.Cache
}

func (m *memory) Init(api core.PluginAPI) error {
	m.pluginApi = api
	adapterConf := &config{}
	if err := mapstructure.Decode(m.rawConf.Config, adapterConf); err != nil {
		return err
	}
	adapterConf.setDefaults()
	m.conf = adapterConf
	m.cache = freecache.NewCache(m.conf.Size * 1024 * 1024)
	return nil
}

func (m *memory) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: m.rawConf.Name,
		ID:   pluginID,
	}
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
