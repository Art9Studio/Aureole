package redis

import (
	storageT "aureole/internal/plugins/storage/types"
	"aureole/plugins/storage"
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
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

func createStorage() (storageT.Storage, error) {
	s := &Storage{client: redis.NewClient(&redis.Options{DB: 15})}
	return s, s.client.Ping(context.Background()).Err()
}
