package etcd

import (
	"aureole/internal/plugins"
	"aureole/plugins/storage"
	"context"
	"errors"
	"go.etcd.io/etcd/clientv3"
	"testing"
	"time"
)

func TestStorage(t *testing.T) {
	s, err := createStorage()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	storage.TestStore(s, t)
	storage.TestTypes(s, t)
}

func createStorage() (plugins.Storage, error) {
	var err error
	s := &etcd{}
	s.client, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Duration(5) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	s.timeout = time.Duration(1) * time.Second

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := s.client.Status(ctxWithTimeout, "127.0.0.1:2379")
	if err != nil {
		return nil, err
	} else if resp == nil {
		return nil, errors.New("the status response from etcd was nil")
	}
	return s, nil
}