package etcd

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/mitchellh/mapstructure"
	"go.etcd.io/etcd/clientv3"
)

const pluginID = "4109"

type etcd struct {
	pluginApi core.PluginAPI
	rawConf   *configs.Storage
	conf      *config
	client    *clientv3.Client
	timeout   time.Duration
}

func (e *etcd) Init(api core.PluginAPI) (err error) {
	e.pluginApi = api
	adapterConf := &config{}
	if err := mapstructure.Decode(e.rawConf.Config, adapterConf); err != nil {
		return err
	}
	adapterConf.setDefaults()
	e.conf = adapterConf
	e.timeout = time.Duration(e.conf.Timeout) * time.Second

	e.client, err = clientv3.New(clientv3.Config{
		Endpoints:   e.conf.Endpoints,
		DialTimeout: time.Duration(e.conf.DialTimeout) * time.Second,
	})
	if err != nil {
		return err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := e.client.Status(ctxWithTimeout, e.conf.Endpoints[0])
	if err != nil {
		return err
	} else if resp == nil {
		return errors.New("the status response from etcd was nil")
	}
	return nil
}

func (e *etcd) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		Name: e.rawConf.Name,
		ID:   pluginID,
	}
}

func (e *etcd) Set(k string, v interface{}, exp int) error {
	if k == "" || v == nil {
		return errors.New("etcd key storage: key and value cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if exp > 0 {
		resp, err := e.client.Grant(context.TODO(), int64(exp))
		if err != nil {
			return err
		}
		_, err = e.client.Put(ctx, k, string(data), clientv3.WithLease(resp.ID))
		return err
	}

	_, err = e.client.Put(ctx, k, string(data))
	return err
}

func (e *etcd) Get(k string, v interface{}) (ok bool, err error) {
	if k == "" || v == nil {
		return false, errors.New("etcd key storage: key and value cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	resp, err := e.client.Get(ctx, k)
	if err != nil {
		return false, err
	}

	if len(resp.Kvs) == 0 {
		return false, nil
	}
	return true, json.Unmarshal(resp.Kvs[0].Value, v)
}

func (e *etcd) Delete(k string) error {
	if k == "" {
		return errors.New("etcd key storage: key and value cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	_, err := e.client.Delete(ctx, k)
	return err
}

func (e *etcd) Exists(k string) (bool, error) {
	if k == "" {
		return false, errors.New("etcd key storage: key and value cannot be empty")
	}
	return e.Get(k, new(interface{}))
}

func (e *etcd) Close() error {
	return e.client.Close()
}
