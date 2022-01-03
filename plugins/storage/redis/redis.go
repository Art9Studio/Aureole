package redis

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"context"
	"encoding/json"
	"errors"
	redisv8 "github.com/go-redis/redis/v8"
	"github.com/mitchellh/mapstructure"
	"time"
)

const pluginID = "5979"

type redis struct {
	pluginApi core.PluginAPI
	rawConf   *configs.Storage
	conf      *config
	client    *redisv8.Client
}

func (s *redis) Init(api core.PluginAPI) error {
	s.pluginApi = api
	adapterConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, adapterConf); err != nil {
		return err
	}
	adapterConf.setDefaults()
	s.conf = adapterConf

	s.client = redisv8.NewClient(&redisv8.Options{
		Addr:     s.conf.Address,
		Password: s.conf.Password,
		DB:       s.conf.DB,
	})
	return s.client.Ping(context.Background()).Err()
}

func (s *redis) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: s.rawConf.Name,
		ID:   pluginID,
	}
}

func (s *redis) Set(k string, v interface{}, exp int) error {
	if k == "" || v == nil {
		return errors.New("redis key storage: key and value cannot be empty")
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = s.client.Set(context.Background(), k, string(data), time.Duration(exp)*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *redis) Get(k string, v interface{}) (ok bool, err error) {
	if k == "" || v == nil {
		return false, errors.New("redis key storage: key and value cannot be empty")
	}

	data, err := s.client.Get(context.Background(), k).Result()
	if err != nil {
		if err == redisv8.Nil {
			return false, nil
		}
		return false, err
	}
	return true, json.Unmarshal([]byte(data), v)
}

func (s *redis) Delete(k string) error {
	if k == "" {
		return errors.New("redis key storage: key and value cannot be empty")
	}

	_, err := s.client.Del(context.Background(), k).Result()
	return err
}

func (s *redis) Exists(k string) (bool, error) {
	if k == "" {
		return false, errors.New("redis key storage: key and value cannot be empty")
	}

	if exists, err := s.client.Exists(context.Background(), k).Result(); err != nil {
		return false, err
	} else {
		return exists == 1, nil
	}
}

func (s *redis) Close() error {
	return s.client.Close()
}
