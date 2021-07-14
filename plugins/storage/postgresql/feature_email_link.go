package postgresql

import (
	"aureole/internal/collections"
	"aureole/internal/plugins/storage/types"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
)

func (s *Storage) InsertEmailLink(spec *collections.Spec, data *types.EmailLinkData) (types.JSONCollResult, error) {
	b := sqlbuilder.PostgreSQL.NewInsertBuilder()

	b.InsertInto(Sanitize(spec.Name))
	b.Cols(Sanitize(spec.FieldsMap["email"].Name),
		Sanitize(spec.FieldsMap["token"].Name),
		Sanitize(spec.FieldsMap["expires"].Name),
		Sanitize(spec.FieldsMap["invalid"].Name))
	b.Values(data.Email, data.Token, data.Expires, data.Invalid)
	b.SQL(fmt.Sprintf(" returning %s", Sanitize(spec.Pk)))

	sql, args := b.Build()
	return s.RawQuery(sql, args...)
}

func (s *Storage) GetEmailLink(spec *collections.Spec, filters []types.Filter) (types.JSONCollResult, error) {
	from := sqlbuilder.PostgreSQL.NewSelectBuilder()
	from.Select(Sanitize(spec.FieldsMap["email"].Name),
		Sanitize(spec.FieldsMap["token"].Name),
		Sanitize(spec.FieldsMap["expires"].Name),
		Sanitize(spec.FieldsMap["invalid"].Name))
	from.From(Sanitize(spec.Name))

	for _, f := range filters {
		from.Where(from.Equal(Sanitize(f.Name), f.Value))
	}

	b := sqlbuilder.PostgreSQL.NewSelectBuilder()
	b.Select("row_to_json(t)")
	b.From(b.BuilderAs(from, "t"))

	sql, args := b.Build()
	return s.RawQuery(sql, args...)
}

func (s *Storage) InvalidateEmailLink(spec *collections.Spec, filters []types.Filter) error {
	b := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	b.Update(Sanitize(spec.Name)).Set(b.Assign(Sanitize(spec.FieldsMap["invalid"].Name), true))

	for _, f := range filters {
		b.Where(b.Equal(Sanitize(f.Name), f.Value))
	}

	b.SQL(fmt.Sprintf(" returning %s", Sanitize(spec.Pk)))
	sql, args := b.Build()

	return s.RawExec(sql, args...)
}
