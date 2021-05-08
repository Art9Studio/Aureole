package postgresql

import (
	"aureole/internal/collections"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	"aureole/internal/plugins/storage/types"
	"fmt"
	"strings"
)

func (s *Storage) CreatePwBasedColl(coll *collections.Collection) error {
	pluginApi := authn.Repository.PluginApi
	identityCollName := coll.Parent
	identityColl, err := pluginApi.Project.GetCollection(identityCollName)
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("alter table %s add column %s text", identityColl.Spec.Name, coll.Spec.FieldsMap["password"])
	return s.RawExec(sql)
}

func (s *Storage) InsertPwBased(i *identity.Identity, identityData *types.IdentityData, pwColl *collections.Collection, pwBasedData *types.PwBasedData) (types.JSONCollResult, error) {
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

	columns = append(columns, pwColl.Spec.FieldsMap["password"])
	data = append(data, pwBasedData.Password)
	values := " values ("

	for i := range columns {
		sql += fmt.Sprintf("%s,", Sanitize(columns[i]))
		values += fmt.Sprintf("$%d,", i+1)
	}

	sql = strings.TrimRight(sql, ",") + ")"
	values = strings.TrimRight(values, ",") + ")"
	sql += values + fmt.Sprintf(" returning %s;", spec.Pk)

	return s.RawQuery(sql, data...)
}

func (s *Storage) GetPassword(coll *collections.Collection, fieldName string, fieldValue interface{}) (types.JSONCollResult, error) {
	pluginApi := authn.Repository.PluginApi
	identityCollName := coll.Parent
	identityColl, err := pluginApi.Project.GetCollection(identityCollName)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf("select %s from %s where %s=$1;",
		Sanitize(coll.Spec.FieldsMap["password"]),
		Sanitize(identityColl.Spec.Name),
		Sanitize(identityColl.Spec.FieldsMap[fieldName]),
	)
	return s.RawQuery(sql, fieldValue)
}
