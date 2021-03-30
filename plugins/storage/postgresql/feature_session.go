package postgresql

import (
	coll "aureole/internal/collections"
	"fmt"
	"time"
)

func (s *Storage) SetGCInterval(interval time.Duration) {
	s.gcInterval = interval
}

func (s *Storage) CreateSessionColl(spec coll.Specification) error {
	// TODO: check types of fields
	sql := fmt.Sprintf(`create table %s
                       (%s text primary key not null default '',
                       %s text not null unique,
                       %s bigint not null default '0');`,
		Sanitize(spec.Name),
		Sanitize(spec.Pk),
		Sanitize(spec.FieldsMap["session_id"]),
		Sanitize(spec.FieldsMap["expiration"]))
	return s.RawExec(sql)
}

func (s *Storage) Get(spec coll.Specification, key string) ([]byte, error) {
	panic("implement me")
}

func (s *Storage) Set(spec coll.Specification, key string, val []byte, exp time.Duration) error {
	panic("implement me")
}

func (s *Storage) Delete(spec coll.Specification, key string) error {
	panic("implement me")
}

func (s *Storage) Reset(spec coll.Specification) error {
	panic("implement me")
}
