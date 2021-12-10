package redis

import (
	"aureole/internal/configs"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/mitchellh/mapstructure"
	"time"
)

type Storage struct {
	rawConf *configs.Storage
	conf    *config
	client  *redis.Client
}

func (s *Storage) Init() error {
	adapterConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, adapterConf); err != nil {
		return err
	}
	adapterConf.setDefaults()
	s.conf = adapterConf

	s.client = redis.NewClient(&redis.Options{
		Addr:     s.conf.Address,
		Password: s.conf.Password,
		DB:       s.conf.DB,
	})
	return s.client.Ping(context.Background()).Err()
}

func (s *Storage) Set(k string, v interface{}, exp int) error {
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

func (s *Storage) Get(k string, v interface{}) (ok bool, err error) {
	if k == "" || v == nil {
		return false, errors.New("redis key storage: key and value cannot be empty")
	}

	data, err := s.client.Get(context.Background(), k).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return true, json.Unmarshal([]byte(data), v)
}

func (s *Storage) Delete(k string) error {
	if k == "" {
		return errors.New("redis key storage: key and value cannot be empty")
	}

	_, err := s.client.Del(context.Background(), k).Result()
	return err
}

func (s *Storage) Exists(k string) (bool, error) {
	if k == "" {
		return false, errors.New("redis key storage: key and value cannot be empty")
	}

	if exists, err := s.client.Exists(context.Background(), k).Result(); err != nil {
		return false, err
	} else {
		return exists == 1, nil
	}
}

func (s *Storage) Close() error {
	return s.client.Close()
}
