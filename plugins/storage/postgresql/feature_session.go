package postgresql

import (
	"aureole/internal/collections"
	"aureole/internal/plugins/storage/types"
)

func (s *Storage) CreateSessionColl(specification collections.Specification) error {
	panic("implement me")
}

func (s *Storage) InsertSession(specification collections.Specification, data types.InsertSessionData) (types.JSONCollResult, error) {
	panic("implement me")
}

func (s *Storage) GetSessionId(specification collections.Specification, i interface{}) (types.JSONCollResult, error) {
	panic("implement me")
}
