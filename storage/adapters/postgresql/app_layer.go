package postgresql

import (
	"fmt"
	"gouth/storage"
)

// IsCollExists checks whether the given collection exists
func (s *Session) IsCollExists(collConf storage.CollConfig) (bool, error) {
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
func (s *Session) CreateUserColl(collConf storage.UserCollConfig) error {
	// TODO: check types of fields
	return s.RawExec(`create table $1
		($2 serial primary key,
		$3 text not null unique,
		$4 text not null);`,
		collConf.Name,
		collConf.Pk,
		collConf.UserUnique,
		collConf.UserConfirm,
	)
}

// InsertUser inserts user entity in the user collection
func (s *Session) InsertUser(collConf storage.UserCollConfig, insUserData storage.InsertUserData) (storage.JSONCollResult, error) {
	return s.RawQuery("insert into $1 ($2, $3) values ($4, $5) returning $6;",
		collConf.Name,
		collConf.UserUnique, collConf.UserConfirm,
		insUserData.UserUnique, insUserData.UserConfirm,
		collConf.Pk,
	)
}

func (s *Session) GetUserPassword() error {
	panic("")
}
