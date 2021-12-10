package memory

import (
	"aureole/internal/configs"
	"encoding/json"
	"errors"
	"github.com/coocood/freecache"
	"github.com/mitchellh/mapstructure"
)

type Storage struct {
	rawConf *configs.Storage
	conf    *config
	cache   *freecache.Cache
}

func (s *Storage) Init() error {
	adapterConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, adapterConf); err != nil {
		return err
	}
	adapterConf.setDefaults()
	s.conf = adapterConf
	s.cache = freecache.NewCache(s.conf.Size * 1024 * 1024)
	return nil
}

func (s *Storage) Set(k string, v interface{}, exp int) error {
	if k == "" || v == nil {
		return errors.New("memory key storage: key and value cannot be empty")
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return s.cache.Set([]byte(k), data, exp)
}

func (s *Storage) Get(k string, v interface{}) (ok bool, err error) {
	if k == "" || v == nil {
		return false, errors.New("memory key storage: key and value cannot be empty")
	}

	data, err := s.cache.Get([]byte(k))
	if err != nil {
		if err == freecache.ErrNotFound {
			return false, nil
		}
		return false, err
	}

	return true, json.Unmarshal(data, v)
}

func (s *Storage) Delete(k string) error {
	if k == "" {
		return errors.New("memory key storage: key and value cannot be empty")
	}
	s.cache.Del([]byte(k))
	return nil
}

func (s *Storage) Exists(k string) (bool, error) {
	if k == "" {
		return false, errors.New("memory key storage: key and value cannot be empty")
	}
	return s.Get(k, new(interface{}))
}

func (s *Storage) Close() error {
	s.cache.Clear()
	return nil
}
