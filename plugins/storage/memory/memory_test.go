package memory

import (
	"aureole/plugins/storage"
	"github.com/coocood/freecache"
	"testing"
)

func TestStorage(t *testing.T) {
	s := &Storage{cache: freecache.NewCache(5 * 1024 * 1024)}
	defer s.Close()
	storage.TestStore(s, t)
	storage.TestTypes(s, t)
}
