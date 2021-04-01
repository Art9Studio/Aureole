package postgresql

import (
	coll "aureole/internal/collections"
	"aureole/internal/plugins/storage/types"
	"context"
	"fmt"
	"time"
)

func (s *Storage) SetCleanInterval(interval time.Duration) {
	s.gcInterval = interval
}

func (s *Storage) CreateSessionColl(spec coll.Specification) error {
	// TODO: check types of fields
	sql := fmt.Sprintf(`create table %s
                       (%s int primary key not null,
                       %s text not null unique,
                       %s bigint not null default '0');`,
		Sanitize(spec.Name),
		Sanitize(spec.Pk),
		Sanitize(spec.FieldsMap["session_token"]),
		Sanitize(spec.FieldsMap["expiration"]))
	return s.RawExec(sql)
}

func (s *Storage) GetSession(spec coll.Specification, userId int) (types.JSONCollResult, error) {
	sql := fmt.Sprintf(`SELECT %s, %s FROM %s WHERE %s=$1;`,
		Sanitize(spec.FieldsMap["session_token"]),
		Sanitize(spec.FieldsMap["expiration"]),
		Sanitize(spec.Name),
		Sanitize(spec.Pk))

	return s.RawQuery(sql, userId)
}

func (s *Storage) InsertSession(spec coll.Specification, data types.InsertSessionData) (types.JSONCollResult, error) {
	expires := data.Expiration.Unix()
	sql := fmt.Sprintf("INSERT INTO %s (%s, %s, %s) VALUES ($1, $2, $3) ON CONFLICT (%s) DO UPDATE SET %s = $4, %s = $5 RETURNING $6",
		Sanitize(spec.Name),
		Sanitize(spec.Pk),
		Sanitize(spec.FieldsMap["session_token"]),
		Sanitize(spec.FieldsMap["expiration"]),
		Sanitize(spec.Pk),
		Sanitize(spec.FieldsMap["session_token"]),
		Sanitize(spec.FieldsMap["expiration"]))
	return s.RawQuery(sql, data.UserId, data.SessionToken, expires, data.SessionToken, expires, spec.Pk)
}

func (s *Storage) DeleteSession(spec coll.Specification, userId int) (types.JSONCollResult, error) {
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s=$1",
		Sanitize(spec.Name),
		Sanitize(spec.Pk))
	return s.RawQuery(sql, userId)
}

func (s *Storage) StartCleaning(spec coll.Specification) {
	go s.gcTicker(spec)
}

// gcTicker starts the gc ticker
func (s *Storage) gcTicker(spec coll.Specification) {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.gcDone:
			return
		case t := <-ticker.C:
			sql := fmt.Sprintf("DELETE FROM %s WHERE %s <= $1 AND %s != 0",
				Sanitize(spec.Name),
				Sanitize(spec.FieldsMap["expiration"]),
				Sanitize(spec.FieldsMap["expiration"]))
			_, _ = s.conn.Exec(context.Background(), sql, t.Unix())
		}
	}
}
