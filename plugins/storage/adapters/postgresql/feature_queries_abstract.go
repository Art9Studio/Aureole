package postgresql

import "aureole/plugins/storage"

func (s *Storage) RelInfo() map[storage.CollPair]storage.RelInfo {
	return s.relInfo
}

func (s *Storage) Read(string) (storage.JSONCollResult, error) {
	panic("implement me")
}
