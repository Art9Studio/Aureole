package postgresql

import (
	"aureole/internal/plugins/storage/types"
)

func (s *Storage) RelInfo() map[types.CollPair]types.RelInfo {
	return s.relInfo
}

func (s *Storage) Read(string) (types.JSONCollResult, error) {
	panic("implement me")
}
