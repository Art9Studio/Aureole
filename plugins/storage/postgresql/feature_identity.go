package postgresql

import (
	"aureole/internal/identity"
	"aureole/internal/plugins/storage/types"
	"fmt"
	"strings"
)

// CreateIdentityColl creates user collection with traits passed by UserCollectionConfig
func (s *Storage) CreateIdentityColl(i *identity.Identity) error {
	// TODO: check types of fields
	spec := i.Collection.Spec
	sql := fmt.Sprintf("create table %s (%s serial primary key",
		Sanitize(spec.Name),
		Sanitize(spec.Pk))

	sql += createField(&i.Username, spec.FieldsMap["username"])
	sql += createField(&i.Phone, spec.FieldsMap["phone"])
	sql += createField(&i.Email, spec.FieldsMap["email"])

	sql += ");"

	return s.RawExec(sql)
}

func createField(field *identity.Trait, fieldName string) (sql string) {
	if field.Enabled {
		sql += fmt.Sprintf(",\n%s text", Sanitize(fieldName))

		if field.Unique {
			sql += " unique"
		}

		if field.Required {
			sql += " not null"
		}
	}

	return sql
}

// InsertIdentity inserts user entity in the user collection
func (s *Storage) InsertIdentity(i *identity.Identity, identityData *types.IdentityData) (types.JSONCollResult, error) {
	spec := i.Collection.Spec
	sql := fmt.Sprintf("insert into %s (", Sanitize(spec.Name))

	var (
		columns []string
		data    []interface{}
	)

	if i.Username.Enabled {
		columns = append(columns, spec.FieldsMap["username"])
		data = append(data, identityData.Username)
	}

	if i.Phone.Enabled {
		columns = append(columns, spec.FieldsMap["phone"])
		data = append(data, identityData.Phone)
	}

	if i.Email.Enabled {
		columns = append(columns, spec.FieldsMap["email"])
		data = append(data, identityData.Email)
	}

	values := " values ("

	for i := range columns {
		sql += fmt.Sprintf("%s,", Sanitize(columns[i]))
		values += fmt.Sprintf("$%d,", i)
	}

	sql = strings.TrimRight(sql, ",") + ")"
	values = strings.TrimRight(values, ",") + ")"
	sql += values + fmt.Sprintf(" returning %s;", spec.Pk)

	return s.RawQuery(sql, data...)
}

func (s *Storage) GetIdentity(i *identity.Identity, fieldName string, fieldValue interface{}) (types.JSONCollResult, error) {
	spec := i.Collection.Spec
	sql := fmt.Sprintf("select row_to_json(t) from ( select %s,", Sanitize(spec.FieldsMap["id"]))

	var columns []string

	if i.Username.Enabled {
		columns = append(columns, spec.FieldsMap["username"])
	}

	if i.Phone.Enabled {
		columns = append(columns, spec.FieldsMap["phone"])
	}

	if i.Email.Enabled {
		columns = append(columns, spec.FieldsMap["email"])
	}

	for i := range columns {
		sql += fmt.Sprintf("%s,", Sanitize(columns[i]))
	}

	sql = strings.TrimRight(sql, ",")
	sql += fmt.Sprintf(" from %s where %s=$1) t;", Sanitize(spec.Name), Sanitize(fieldName))

	return s.RawQuery(sql, fieldValue)
}
