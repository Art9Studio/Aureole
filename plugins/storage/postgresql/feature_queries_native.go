package postgresql

import (
	"aureole/internal/plugins/storage/types"
)

func (s *Storage) NativeQuery(query string, args ...interface{}) (types.JSONCollResult, error) {
	// todo: think about json aggregation func
	// sql := fmt.Sprintf("select json_agg(t) from (%s) t", query)
	return s.RawQuery(query, args...)
}
