package postgresql

import "gouth/storage"

func (s *Session) RelInfo() map[storage.CollPair]storage.RelInfo {
	return s.relInfo
}

func (s *Session) Read(string) (storage.JSONCollResult, error) {
	panic("implement me")
}
