package postgresql

import (
	"aureole/internal/collections"
	"aureole/internal/identity"
	"aureole/internal/plugins/authn"
	"aureole/internal/plugins/storage/types"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
	"time"
)

func (s *Storage) InsertPwBased(i *identity.Identity, pwColl *collections.Collection, iData *types.IdentityData, pwData *types.PwBasedData) (types.JSONCollResult, error) {
	var (
		cols   []string
		values []interface{}
	)

	spec := i.Collection.Spec
	if iData.Username != nil {
		cols = append(cols, Sanitize(spec.FieldsMap["username"].Name))
		values = append(values, iData.Username)
	}

	if iData.Phone != nil {
		cols = append(cols, Sanitize(spec.FieldsMap["phone"].Name))
		values = append(values, iData.Phone)
	}

	if iData.Email != nil {
		cols = append(cols, Sanitize(spec.FieldsMap["email"].Name))
		values = append(values, iData.Email)
	}

	for fieldName := range i.Additional {
		cols = append(cols, Sanitize(spec.FieldsMap[fieldName].Name))
		values = append(values, iData.Additional[fieldName])
	}

	if created := spec.FieldsMap["created"]; created.Name != "" {
		cols = append(cols, Sanitize(created.Name))
		values = append(values, time.Now())
	}

	if isActive := spec.FieldsMap["is_active"]; isActive.Name != "" {
		cols = append(cols, Sanitize(isActive.Name))
		values = append(values, true)
	}

	if emailVerif := spec.FieldsMap["email_verified"]; emailVerif.Name != "" {
		cols = append(cols, Sanitize(emailVerif.Name))
		values = append(values, false)
	}

	if phoneVerif := spec.FieldsMap["phone_verified"]; phoneVerif.Name != "" {
		cols = append(cols, Sanitize(phoneVerif.Name))
		values = append(values, false)
	}

	cols = append(cols, pwColl.Spec.FieldsMap["password"].Name)
	values = append(values, pwData.PasswordHash)

	builder := sqlbuilder.PostgreSQL.NewInsertBuilder()
	builder.InsertInto(Sanitize(spec.Name))
	builder.Cols(cols...).Values(values...).SQL(fmt.Sprintf(" returning %s", Sanitize(spec.Pk)))
	sql, args := builder.Build()

	return s.RawQuery(sql, args...)
}

func (s *Storage) GetPassword(coll *collections.Collection, filters []types.Filter) (types.JSONCollResult, error) {
	pluginApi := authn.Repository.PluginApi
	identityColl, err := pluginApi.Project.GetCollection(coll.ParentName)
	if err != nil {
		return nil, err
	}

	b := sqlbuilder.PostgreSQL.NewSelectBuilder()
	b.Select(Sanitize(coll.Spec.FieldsMap["password"].Name)).From(Sanitize(identityColl.Spec.Name))

	for _, f := range filters {
		b.Where(b.Equal(Sanitize(f.Name), f.Value))
	}

	sql, args := b.Build()

	return s.RawQuery(sql, args...)
}

func (s *Storage) UpdatePassword(coll *collections.Collection, filters []types.Filter, newPw interface{}) (types.JSONCollResult, error) {
	pluginApi := authn.Repository.PluginApi
	identityColl, err := pluginApi.Project.GetCollection(coll.ParentName)
	if err != nil {
		return nil, err
	}

	b := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	b.Update(Sanitize(identityColl.Spec.Name)).Set(b.Assign(Sanitize(coll.Spec.FieldsMap["password"].Name), newPw))

	for _, f := range filters {
		b.Where(b.Equal(Sanitize(f.Name), f.Value))
	}

	b.SQL(fmt.Sprintf(" returning %s", Sanitize(identityColl.Spec.Pk)))
	sql, args := b.Build()

	return s.RawQuery(sql, args...)
}

/* Funcs for creating table from scratch. Enables by py passing "use_existent: false" flag

func (s *Storage) CreatePwBasedColl(coll *collections.Collection) error {
	pluginApi := authn.Repository.PluginApi
	identityColl, err := pluginApi.Project.GetCollection(coll.ParentName)
	if err != nil {
		return err
	}

	var fieldType string
	pwSpec := coll.Spec.FieldsMap["password"]
	if pwSpec.Type != "" {
		fieldType = pwSpec.Type
	} else {
		fieldType = DefaultFieldType
	}

	sql := fmt.Sprintf("alter table %s add column %s %s", identityColl.Spec.Name, pwSpec.Name, fieldType)
	return s.RawExec(sql)
}
*/
