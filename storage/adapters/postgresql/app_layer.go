package postgresql

import (
	"fmt"
	"gouth/storage"
)

// IsCollExists checks whether the given collection exists
func (s *Session) IsCollExists(collConf storage.CollectionConfig) (bool, error) {
	sql := fmt.Sprintf(
		"select exists(select res from (select to_regclass('%s')) as res where res is not null);",
		collConf.Name())
	res, err := s.RawQuery(sql)

	if err != nil {
		return false, err
	}

	return res.(bool), nil
}

// CreateUserCollection creates user collection with traits passed by UserCollectionConfig
func (s *Session) CreateUserColl(collConf storage.UserCollConfig) error {
	// TODO: check types of fields
	sql := fmt.Sprintf(
		`create table %s
		(%s serial primary key,
		%s varchar(50) not null unique,
		%s varchar(50) not null);`,
		collConf.Name(),
		collConf.PK(),
		collConf.UserID(),
		collConf.UserConfirm(),
	)

	if err := s.RawExec(sql); err != nil {
		return err
	}

	return nil
}

// InsertUser inserts user entity in the user collection
func (s *Session) InsertUser(collConf storage.UserCollConfig, insUserData storage.InsertionUserData) (storage.JSONCollResult, error) {
	// TODO: make possible to be UserID not only string
	sql := fmt.Sprintf(
		"insert into %s (%s, %s) values ('%s', '%s') returning %s;",
		collConf.Name(),
		collConf.UserID(), collConf.UserConfirm(),
		insUserData.UserID(), insUserData.UserConfirm(),
		collConf.PK(),
	)

	return s.RawQuery(sql)
}

func (s *Session) GetUserPassword() error {
	panic("")
}
