package postgresql

import (
	"fmt"
	"github.com/jackc/pgx/v4"
	"gouth/adapters/storage"
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
func (s *ConnSession) CreateIdentityColl(collConf storage.IdentityCollConfig) error {
	// TODO: check types of fields
	sql := fmt.Sprintf(`create table %s
                       (%s serial primary key,
                       %s text not null unique,
                       %s text not null);`,
		Sanitize(collConf.Name),
		Sanitize(collConf.Pk),
		Sanitize(collConf.Identity),
		Sanitize(collConf.Password))
	return s.RawExec(sql)
}

// InsertUser inserts user entity in the user collection
func (s *ConnSession) InsertIdentity(collConf storage.IdentityCollConfig, insUserData storage.InsertUserData) (storage.JSONCollResult, error) {
	sql := fmt.Sprintf("insert into %s (%s, %s) values ($1, $2) returning $3;",
		Sanitize(collConf.Name),
		Sanitize(collConf.Identity),
		Sanitize(collConf.Password))
	return s.RawQuery(sql, insUserData.Identity, insUserData.UserConfirm, collConf.Pk)
}

func (s *ConnSession) GetPasswordByIdentity(collConf storage.IdentityCollConfig, userUnique interface{}) (storage.JSONCollResult, error) {
	sql := fmt.Sprintf("select %s from %s where %s=$1",
		Sanitize(collConf.Password),
		Sanitize(collConf.Name),
		Sanitize(collConf.Identity),
	)
	return s.RawQuery(sql, userUnique)
}

func Sanitize(ident string) string {
	return pgx.Identifier.Sanitize([]string{ident})
}
