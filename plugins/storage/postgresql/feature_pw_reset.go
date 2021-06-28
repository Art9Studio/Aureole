package postgresql

import (
	"aureole/internal/collections"
	"aureole/internal/plugins/storage/types"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
)

func (s *Storage) InsertReset(spec *collections.Spec, data *types.PwResetData) (types.JSONCollResult, error) {
	b := sqlbuilder.PostgreSQL.NewInsertBuilder()

	b.InsertInto(Sanitize(spec.Name))
	b.Cols(Sanitize(spec.FieldsMap["email"].Name),
		Sanitize(spec.FieldsMap["token"].Name),
		Sanitize(spec.FieldsMap["expires"].Name))
	b.Values(data.Email, data.Token, data.Expires)
	b.SQL(fmt.Sprintf(" returning %s", Sanitize(spec.Pk)))

	sql, args := b.Build()
	return s.RawQuery(sql, args...)
}

func (s *Storage) GetReset(spec *collections.Spec, filterField string, filterValue interface{}) (types.JSONCollResult, error) {
	from := sqlbuilder.PostgreSQL.NewSelectBuilder()
	from.Select(Sanitize(spec.FieldsMap["email"].Name),
		Sanitize(spec.FieldsMap["token"].Name),
		Sanitize(spec.FieldsMap["expires"].Name))
	from.From(Sanitize(spec.Name)).Where(from.Equal(Sanitize(filterField), filterValue))

	b := sqlbuilder.PostgreSQL.NewSelectBuilder()
	b.Select("row_to_json(t)")
	b.From(b.BuilderAs(from, "t"))

	sql, args := b.Build()
	return s.RawQuery(sql, args...)
}
