package postgresql

import (
	"aureole/internal/plugins/storage/types"
	"fmt"
)

func (s *Storage) NativeQuery(query string, args ...interface{}) (types.JSONCollResult, error) {
	sql := fmt.Sprintf("select json_agg(t) from (%s) t", query)
	return s.RawQuery(sql, args...)
}
