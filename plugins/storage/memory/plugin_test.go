package memory

import (
	"aureole/plugins/storage"
	"testing"

	"github.com/coocood/freecache"
)

func TestStorage(t *testing.T) {
	s := &memory{cache: freecache.NewCache(5 * 1024 * 1024)}
	defer s.Close()
	storage.TestStore(s, t)
	storage.TestTypes(s, t)
}
