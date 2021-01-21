package postgresql

import (
	"fmt"
	adapters "gouth/storage"
)

type UserCollectionConfig struct {
	collection  string
	pk          string
	userId      string
	userConfirm string
}

func (u UserCollectionConfig) Collection() string {
	return u.collection
}

func (u UserCollectionConfig) Pk() string {
	return u.pk
}

func (u UserCollectionConfig) UserId() string {
	return u.userId
}

func (u UserCollectionConfig) UserConfirm() string {
	return u.userConfirm
}

type InsertUserData struct {
	userId      string
	userConfirm string
}

func (i InsertUserData) UserId() string {
	return i.userId
}

func (i InsertUserData) UserConfirm() string {
	return i.userConfirm
}

// IsUserCollectionExists checks whether the given collection exists
func (s *Session) IsUserCollectionExists(colConf adapters.UserCollectionConfig) (bool, error) {
	sql := fmt.Sprintf(
		"select exists(select res from (select to_regclass('%s')) as res where res is not null);",
		colConf.Collection())
	res, err := s.RawQuery(sql)

	if err != nil {
		return false, err
	}

	return res.(bool), nil
}

// CreateUserCollection creates user collection with traits passed by UserCollectionConfig
func (s *Session) CreateUserCollection(colConf adapters.UserCollectionConfig) error {
	// todo: check types of fields
	sql := fmt.Sprintf(
		`create table %s
		(%s serial primary key,
		%s varchar(50) not null unique,
		%s varchar(50) not null);`,
		colConf.Collection(),
		colConf.Pk(),
		colConf.UserId(),
		colConf.UserConfirm(),
	)

	if err := s.RawExec(sql); err != nil {
		return err
	}

	return nil
}

// InsertUser inserts user entity in the user collection
func (s *Session) InsertUser(colConf adapters.UserCollectionConfig, insUsrConf adapters.InsertUserData) (adapters.JSONCollectionResult, error) {
	// todo: make possible to be UserId not only string
	sql := fmt.Sprintf(
		"insert into %s (%s, %s) values ('%s', '%s') returning %s;",
		colConf.Collection(),
		colConf.UserId(), colConf.UserConfirm(),
		insUsrConf.UserId(), insUsrConf.UserConfirm(),
		colConf.Pk(),
	)

	return s.RawQuery(sql)
}

func (s *Session) GetUserPassword() error {
	panic("")
}
