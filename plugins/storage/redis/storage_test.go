package redis

import (
	"aureole/internal/plugins"
	"aureole/plugins/storage"
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

func createStorage() (plugins.Storage, error) {
	s := &storage{client: redisv8.NewClient(&redisv8.Options{DB: 15})}
	return s, s.client.Ping(context.Background()).Err()
}
