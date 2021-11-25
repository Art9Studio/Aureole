package etcd

import (
	"aureole/internal/configs"
	"aureole/internal/plugins"
	"aureole/internal/plugins/core"
	"context"
	"encoding/json"
	"errors"
	"github.com/mitchellh/mapstructure"
	"go.etcd.io/etcd/clientv3"
	"time"
)

const PluginID = "4109"

type Storage struct {
	pluginApi core.PluginAPI
	rawConf   *configs.Storage
	conf      *config
	client    *clientv3.Client
	timeout   time.Duration
}

func (s *Storage) Init(api core.PluginAPI) (err error) {
	s.pluginApi = api
	adapterConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, adapterConf); err != nil {
		return err
	}
	adapterConf.setDefaults()
	s.conf = adapterConf
	s.timeout = time.Duration(s.conf.Timeout) * time.Second

	s.client, err = clientv3.New(clientv3.Config{
		Endpoints:   s.conf.Endpoints,
		DialTimeout: time.Duration(s.conf.DialTimeout) * time.Second,
	})
	if err != nil {
		return err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := s.client.Status(ctxWithTimeout, s.conf.Endpoints[0])
	if err != nil {
		return err
	} else if resp == nil {
		return errors.New("the status response from etcd was nil")
	}
	return nil
}

func (s *Storage) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: AdapterName,
		Name: s.rawConf.Name,
		ID:   PluginID,
	}
}

func (s *Storage) Set(k string, v interface{}, exp int) error {
	if k == "" || v == nil {
		return errors.New("etcd key storage: key and value cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if exp > 0 {
		resp, err := s.client.Grant(context.TODO(), int64(exp))
		if err != nil {
			return err
		}
		_, err = s.client.Put(ctx, k, string(data), clientv3.WithLease(resp.ID))
		return err
	}

	_, err = s.client.Put(ctx, k, string(data))
	return err
}

func (s *Storage) Get(k string, v interface{}) (ok bool, err error) {
	if k == "" || v == nil {
		return false, errors.New("etcd key storage: key and value cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := s.client.Get(ctx, k)
	if err != nil {
		return false, err
	}

	if len(resp.Kvs) == 0 {
		return false, nil
	}
	return true, json.Unmarshal(resp.Kvs[0].Value, v)
}

func (s *Storage) Delete(k string) error {
	if k == "" {
		return errors.New("etcd key storage: key and value cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	_, err := s.client.Delete(ctx, k)
	return err
}

func (s *Storage) Exists(k string) (bool, error) {
	if k == "" {
		return false, errors.New("etcd key storage: key and value cannot be empty")
	}
	return s.Get(k, new(interface{}))
}

func (s *Storage) Close() error {
	return s.client.Close()
}
