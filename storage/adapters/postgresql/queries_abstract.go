package postgresql

import adapters "gouth/storage"

func (s *Session) RelInfo() map[adapters.CollectionPair]adapters.RelInfo {
	return s.relInfo
}

func (s *Session) Read(string) (adapters.JSONCollectionResult, error) {
	panic("implement me")
}
