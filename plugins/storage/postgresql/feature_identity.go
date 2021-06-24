package postgresql

import (
	"aureole/internal/collections"
	"aureole/internal/identity"
	"aureole/internal/plugins/storage/types"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
)

// InsertIdentity inserts user entity in the user collection
func (s *Storage) InsertIdentity(i *identity.Identity, iData *types.IdentityData) (types.JSONCollResult, error) {
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
		values = append(values, iData.Additional["is_active"])
	}

	if emailVerif := spec.FieldsMap["email_verified"]; emailVerif.Name != "" {
		cols = append(cols, Sanitize(emailVerif.Name))
		values = append(values, false)
	}

	if phoneVerif := spec.FieldsMap["phone_verified"]; phoneVerif.Name != "" {
		cols = append(cols, Sanitize(phoneVerif.Name))
		values = append(values, false)
	}

	// todo: fix empty row inserting

	b := sqlbuilder.PostgreSQL.NewInsertBuilder()
	b.InsertInto(Sanitize(spec.Name))
	b.Cols(cols...).Values(values...).SQL(fmt.Sprintf(" returning %s", Sanitize(spec.Pk)))
	sql, args := b.Build()

	return s.RawQuery(sql, args...)
}

func (s *Storage) GetIdentity(i *identity.Identity, filters []types.Filter) (types.JSONCollResult, error) {
	var cols []string

	spec := i.Collection.Spec
	if i.Id.IsEnabled {
		cols = append(cols, Sanitize(spec.FieldsMap["id"].Name))
	}

	if i.Username.IsEnabled {
		cols = append(cols, Sanitize(spec.FieldsMap["username"].Name))
	}

	if i.Phone.IsEnabled {
		cols = append(cols, Sanitize(spec.FieldsMap["phone"].Name))
	}

	if i.Email.IsEnabled {
		cols = append(cols, Sanitize(spec.FieldsMap["email"].Name))
	}

	for fieldName := range i.Additional {
		cols = append(cols, Sanitize(spec.FieldsMap[fieldName].Name))
	}

	if created := spec.FieldsMap["created"]; created.Name != "" {
		cols = append(cols, Sanitize(created.Name))
	}

	if isActive := spec.FieldsMap["is_active"]; isActive.Name != "" {
		cols = append(cols, Sanitize(isActive.Name))
	}

	if emailVerif := spec.FieldsMap["email_verified"]; emailVerif.Name != "" {
		cols = append(cols, Sanitize(emailVerif.Name))
	}

	if phoneVerif := spec.FieldsMap["phone_verified"]; phoneVerif.Name != "" {
		cols = append(cols, Sanitize(phoneVerif.Name))
	}

	from := sqlbuilder.PostgreSQL.NewSelectBuilder()
	from.Select(cols...).From(Sanitize(spec.Name))

	for _, f := range filters {
		from.Where(from.Equal(Sanitize(f.Name), f.Value))
	}

	b := sqlbuilder.PostgreSQL.NewSelectBuilder()
	b.Select("row_to_json(t)")
	b.From(b.BuilderAs(from, "t"))
	sql, args := b.Build()

	return s.RawQuery(sql, args...)
}

func (s *Storage) IsIdentityExist(i *identity.Identity, filters []types.Filter) (bool, error) {
	spec := i.Collection.Spec
	q := sqlbuilder.PostgreSQL.NewSelectBuilder()
	q.Select("1").From(Sanitize(spec.Name))

	for _, f := range filters {
		q.Where(q.Equal(Sanitize(f.Name), f.Value))
	}

	b := sqlbuilder.WithFlavor(sqlbuilder.Buildf("SELECT exists (%v)", q), sqlbuilder.PostgreSQL)
	sql, args := b.Build()

	res, err := s.RawQuery(sql, args...)
	if err != nil {
		return false, err
	}

	return res.(bool), nil
}

func (s *Storage) SetEmailVerified(spec *collections.Spec, filters []types.Filter) error {
	b := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	b.Update(Sanitize(spec.Name)).Set(b.Assign(Sanitize(spec.FieldsMap["email_verified"].Name), true))

	for _, f := range filters {
		b.Where(b.Equal(Sanitize(f.Name), f.Value))
	}

	sql, args := b.Build()

	return s.RawExec(sql, args...)
}

func (s *Storage) SetPhoneVerified(spec *collections.Spec, filters []types.Filter) error {
	b := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	b.Update(Sanitize(spec.Name)).Set(b.Assign(Sanitize(spec.FieldsMap["phone_verified"].Name), true))

	for _, f := range filters {
		b.Where(b.Equal(Sanitize(f.Name), f.Value))
	}

	sql, args := b.Build()

	return s.RawExec(sql, args...)
}

/* Funcs for creating table from scratch. Enables by py passing "use_existent: false" flag

// CreateIdentityColl creates user collection with traits passed by UserCollectionConfig
func (s *Storage) CreateIdentityColl(i *identity.Identity) error {
	spec := i.Collection.Spec
	pk := spec.Pk

	builder := sqlbuilder.PostgreSQL.NewCreateTableBuilder()
	builder.CreateTable(Sanitize(spec.Name))
	builder.Define(Sanitize(pk), spec.FieldsMap[pk].Type, "primary key")
	builder.Define(createField(spec.FieldsMap["username"], i.Username.IsUnique, i.Username.IsRequired)...)
	builder.Define(createField(spec.FieldsMap["phone"], i.Phone.IsUnique, i.Phone.IsRequired)...)
	builder.Define(createField(spec.FieldsMap["email"], i.Email.IsUnique, i.Email.IsRequired)...)

	for fieldName, field := range i.Additional {
		builder.Define(createField(spec.FieldsMap[fieldName], field.IsUnique, field.IsRequired)...)
	}
	sql, _ := builder.Build()

	return s.RawExec(sql)
}

func createField(fieldSpec collections.FieldSpec, isUnique, isRequired bool) []string {
	sql := []string{fieldSpec.Name}

	if fieldSpec.Type != "" {
		sql = append(sql, fieldSpec.Type)
	} else {
		sql = append(sql, DefaultFieldType)
	}

	if isUnique {
		sql = append(sql, "unique")
	}

	if isRequired {
		sql = append(sql, "not null")
	}

	return sql
}
*/
