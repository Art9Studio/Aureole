package postgresql

import (
	"aureole/internal/plugins/storage/types"
)

func (s *Storage) NativeQuery(query string, args ...interface{}) (types.JSONCollResult, error) {
	return s.RawQuery(query, args...)
}
