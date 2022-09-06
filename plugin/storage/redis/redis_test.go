package redis

import (
	"aureole/internal/core"
	"aureole/plugin/storage"
	"context"
	"testing"

	redisv8 "github.com/go-redis/redis/v8"
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

func createStorage() (core.Storage, error) {
	s := &redis{client: redisv8.NewClient(&redisv8.Options{DB: 15})}
	return s, s.client.Ping(context.Background()).Err()
}
