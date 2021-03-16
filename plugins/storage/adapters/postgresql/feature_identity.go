package postgresql

import (
	"aureole/collections"
	storageTypes "aureole/plugins/storage/types"
	"fmt"
	"github.com/jackc/pgx/v4"
)

// IsCollExists checks whether the given collection exists
func (s *Storage) IsCollExists(spec collections.Specification) (bool, error) {
	// TODO: use current schema instead constant 'public'
	sql := fmt.Sprintf(
		"select exists (select from pg_tables where schemaname = 'public' AND tablename = '%s');",
		spec.Name)
	res, err := s.RawQuery(sql)

	if err != nil {
		return false, err
	}

	return res.(bool), nil
}

// CreateIdentityColl creates user collection with traits passed by UserCollectionConfig
func (s *Storage) CreateIdentityColl(spec collections.Specification) error {
	// TODO: check types of fields
	sql := fmt.Sprintf(`create table %s
                       (%s serial primary key,
                       %s text not null unique,
                       %s text not null);`,
		Sanitize(spec.Name),
		Sanitize(spec.Pk),
		Sanitize(spec.FieldsMap["identity"]),
		Sanitize(spec.FieldsMap["password"]))
	return s.RawExec(sql)
}

// InsertIdentity inserts user entity in the user collection
func (s *Storage) InsertIdentity(spec collections.Specification, insUserData storageTypes.InsertIdentityData) (storageTypes.JSONCollResult, error) {
	sql := fmt.Sprintf("insert into %s (%s, %s) values ($1, $2) returning $3;",
		Sanitize(spec.Name),
		Sanitize(spec.FieldsMap["identity"]),
		Sanitize(spec.FieldsMap["password"]))
	return s.RawQuery(sql, insUserData.Identity, insUserData.UserConfirm, spec.Pk)
}

func (s *Storage) GetPasswordByIdentity(spec collections.Specification, userUnique interface{}) (storageTypes.JSONCollResult, error) {
	sql := fmt.Sprintf("select %s from %s where %s=$1",
		Sanitize(spec.FieldsMap["password"]),
		Sanitize(spec.Name),
		Sanitize(spec.FieldsMap["identity"]),
	)
	return s.RawQuery(sql, userUnique)
}

func Sanitize(ident string) string {
	return pgx.Identifier.Sanitize([]string{ident})
}
