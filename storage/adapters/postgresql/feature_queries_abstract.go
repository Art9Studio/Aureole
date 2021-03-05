package postgresql

import "aureole/storage"

func (s *ConnSession) RelInfo() map[storage.CollPair]storage.RelInfo {
	return s.relInfo
}

func (s *ConnSession) Read(string) (storage.JSONCollResult, error) {
	panic("implement me")
}
