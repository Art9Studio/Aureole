package postgresql

import (
	"aureole/internal/collections"
	"aureole/internal/plugins/storage/types"
	"fmt"
	"time"
)

func (s *Storage) GetSession(spec collections.Spec, userId int) (types.JSONCollResult, error) {
	sql := fmt.Sprintf(`SELECT %s, %s FROM %s WHERE %s=$1;`,
		Sanitize(spec.FieldsMap["session_token"].Name),
		Sanitize(spec.FieldsMap["expiration"].Name),
		Sanitize(spec.Name),
		Sanitize(spec.Pk))

	return s.RawQuery(sql, userId)
}

func (s *Storage) InsertSession(spec collections.Spec, data types.InsertSessionData) (types.JSONCollResult, error) {
	expires := data.Expiration
	sql := fmt.Sprintf("INSERT INTO %s (%s, %s, %s) VALUES ($1, $2, $3) ON CONFLICT (%s) DO UPDATE SET %s = $4, %s = $5 RETURNING $6",
		Sanitize(spec.Name),
		Sanitize(spec.Pk),
		Sanitize(spec.FieldsMap["session_token"].Name),
		Sanitize(spec.FieldsMap["expiration"].Name),
		Sanitize(spec.Pk),
		Sanitize(spec.FieldsMap["session_token"].Name),
		Sanitize(spec.FieldsMap["expiration"].Name))
	return s.RawQuery(sql, data.UserId, data.SessionToken, expires, data.SessionToken, expires, spec.Pk)
}

func (s *Storage) DeleteSession(spec collections.Spec, userId int) (types.JSONCollResult, error) {
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s=$1",
		Sanitize(spec.Name),
		Sanitize(spec.Pk))
	return s.RawQuery(sql, userId)
}

func (s *Storage) SetCleanInterval(interval int) {
	s.gcInterval = time.Duration(interval) * time.Second
}

func (s *Storage) StartCleaning(spec collections.Spec) {
	go s.cleanTicker(spec)
}

// cleanTicker starts the gc ticker
func (s *Storage) cleanTicker(spec collections.Spec) {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.gcDone:
			return
		case t := <-ticker.C:
			sql := fmt.Sprintf("DELETE FROM %s WHERE %s <= $1 AND %s != 0",
				Sanitize(spec.Name),
				Sanitize(spec.FieldsMap["expiration"].Name),
				Sanitize(spec.FieldsMap["expiration"].Name))
			_ = s.RawExec(sql, t.Unix())
		}
	}
}

/* Func for creating table from scratch

func (s *Storage) CreateSessionColl(spec collections.Spec) error {
	// TODO: check types of fields
	sql := fmt.Sprintf(`create table %s
                       (%s int primary key not null,
                       %s text not null unique,
                       %s bigint not null default '0');`,
		Sanitize(spec.Name),
		Sanitize(spec.Pk),
		Sanitize(spec.FieldsMap["session_token"].Name),
		Sanitize(spec.FieldsMap["expiration"].Name))
	return s.RawExec(sql)
}
*/
