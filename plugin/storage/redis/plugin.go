package redis

import (
	"aureole/configs"
	"aureole/internal/core"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
	"github.com/mitchellh/mapstructure"
)

// const pluginID = "5979"

//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

// init initializes package by register pluginCreator
func init() {
	meta = core.StorageRepo.Register(rawMeta, Create)
}

type redis struct {
	pluginApi core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
	client    *redisv8.Client
}

func Create(conf configs.PluginConfig) core.Storage {
	return &redis{rawConf: conf}
}
func (s *redis) Init(api core.PluginAPI) error {
	s.pluginApi = api
	PluginConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, PluginConf); err != nil {
		return err
	}
	PluginConf.setDefaults()
	s.conf = PluginConf

	s.client = redisv8.NewClient(&redisv8.Options{
		Addr:     s.conf.Address,
		Password: s.conf.Password,
		DB:       s.conf.DB,
	})
	return s.client.Ping(context.Background()).Err()
}

func (s redis) GetMetadata() core.Metadata {
	return meta
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
func (r *redis) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{}
}

func (s *redis) Close() error {
	return s.client.Close()
}
