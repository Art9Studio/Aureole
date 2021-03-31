package postgresql

import (
	coll "aureole/internal/collections"
	"aureole/internal/plugins/storage/types"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
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

func (s *Storage) GetSession(spec coll.Specification, userId int) (types.JSONCollResult, error) {
	sql := fmt.Sprintf(`SELECT %s, %s FROM %s WHERE %s=$1;`,
		Sanitize(spec.FieldsMap["session_id"]),
		Sanitize(spec.FieldsMap["expiration"]),
		Sanitize(spec.Name),
		Sanitize(spec.Pk))

	session, err := s.RawQuery(sql, userId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	rawExp := session.(map[string]interface{})[spec.FieldsMap["expiration"]]
	exp := rawExp.(int64)
	if exp != 0 && exp <= time.Now().Unix() {
		return nil, nil
	}

	return session, nil
}

func (s *Storage) InsertSession(spec coll.Specification, data types.InsertSessionData) (types.JSONCollResult, error) {
	expSeconds := time.Now().Add(data.Expiration).Unix()
	sql := fmt.Sprintf("INSERT INTO %s (%s, %s, %s) VALUES ($1, $2, $3) ON CONFLICT (%s) DO UPDATE SET v = $4, e = $5",
		Sanitize(spec.Name),
		Sanitize(spec.Pk),
		Sanitize(spec.FieldsMap["session_id"]),
		Sanitize(spec.FieldsMap["expiration"]),
		Sanitize(spec.Pk))
	return s.RawQuery(sql, data.UserId, data.SessionToken, expSeconds, data.SessionToken, expSeconds)
}

func (s *Storage) DeleteSession(spec coll.Specification, userId int) (types.JSONCollResult, error) {
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s=$1",
		Sanitize(spec.Name),
		Sanitize(spec.Pk))
	return s.RawQuery(sql, userId)
}

func (s *Storage) StartGC(spec coll.Specification) {
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
