package postgresql

import (
	"aureole/storage"
	"fmt"
	"github.com/jackc/pgx/v4"
)

// IsCollExists checks whether the given collection exists
func (s *ConnSession) IsCollExists(collConf storage.CollConfig) (bool, error) {
	// TODO: use current schema instead constant 'public'
	sql := fmt.Sprintf(
		"select exists (select from pg_tables where schemaname = 'public' AND tablename = '%s');",
		collConf.Name)
	res, err := s.RawQuery(sql)

	if err != nil {
		return false, err
	}

	return res.(bool), nil
}

// CreateUserCollection creates user collection with traits passed by UserCollectionConfig
func (s *ConnSession) CreateUserColl(collConf storage.UserCollConfig) error {
	// TODO: check types of fields
	sql := fmt.Sprintf(`create table %s
                       (%s serial primary key,
                       %s text not null unique,
                       %s text not null);`,
		Sanitize(collConf.Name),
		Sanitize(collConf.Pk),
		Sanitize(collConf.UserUnique),
		Sanitize(collConf.UserConfirm))
	return s.RawExec(sql)
}

// InsertUser inserts user entity in the user collection
func (s *ConnSession) InsertUser(collConf storage.UserCollConfig, insUserData storage.InsertUserData) (storage.JSONCollResult, error) {
	sql := fmt.Sprintf("insert into %s (%s, %s) values ($1, $2) returning $3;",
		Sanitize(collConf.Name),
		Sanitize(collConf.UserUnique),
		Sanitize(collConf.UserConfirm))
	return s.RawQuery(sql, insUserData.UserUnique, insUserData.UserConfirm, collConf.Pk)
}

func (s *ConnSession) GetUserPassword(collConf storage.UserCollConfig, userUnique interface{}) (storage.JSONCollResult, error) {
	sql := fmt.Sprintf("select %s from %s where %s=$1",
		Sanitize(collConf.UserConfirm),
		Sanitize(collConf.Name),
		Sanitize(collConf.UserUnique),
	)
	return s.RawQuery(sql, userUnique)
}

func Sanitize(ident string) string {
	return pgx.Identifier.Sanitize([]string{ident})
}
