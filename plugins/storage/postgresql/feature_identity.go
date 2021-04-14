package postgresql

import (
	"aureole/internal/collections"
	"aureole/internal/plugins/storage/types"
	"fmt"
)

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
func (s *Storage) InsertIdentity(spec collections.Specification, insUserData types.InsertIdentityData) (types.JSONCollResult, error) {
	sql := fmt.Sprintf("insert into %s (%s, %s) values ($1, $2) returning %s;",
		Sanitize(spec.Name),
		Sanitize(spec.FieldsMap["identity"]),
		Sanitize(spec.FieldsMap["password"]),
		Sanitize(spec.Pk))
	return s.RawQuery(sql, insUserData.Identity, insUserData.UserConfirm)
}

func (s *Storage) GetPasswordByIdentity(spec collections.Specification, userUnique interface{}) (types.JSONCollResult, error) {
	sql := fmt.Sprintf("select %s from %s where %s=$1",
		Sanitize(spec.FieldsMap["password"]),
		Sanitize(spec.Name),
		Sanitize(spec.FieldsMap["identity"]),
	)
	return s.RawQuery(sql, userUnique)
}
