package postgresql

import (
	"aureole/internal/collections"
	"aureole/internal/plugins/storage/types"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
)

func (s *Storage) InsertVerification(spec *collections.Spec, confirmData *types.PhoneVerificationData) (types.JSONCollResult, error) {
	b := sqlbuilder.PostgreSQL.NewInsertBuilder()

	b.InsertInto(Sanitize(spec.Name))
	b.Cols(Sanitize(spec.FieldsMap["phone"].Name), Sanitize(spec.FieldsMap["code"].Name),
		Sanitize(spec.FieldsMap["attempts"].Name), Sanitize(spec.FieldsMap["expires"].Name))
	b.Values(confirmData.Phone, confirmData.Code, confirmData.Attempts, confirmData.Expires)
	b.SQL(fmt.Sprintf(" returning %s", Sanitize(spec.Pk)))

	sql, args := b.Build()
	return s.RawQuery(sql, args...)
}

func (s *Storage) GetVerification(spec *collections.Spec, filterField string, filterValue interface{}) (types.JSONCollResult, error) {
	from := sqlbuilder.PostgreSQL.NewSelectBuilder()
	from.Select(Sanitize(spec.FieldsMap["phone"].Name), Sanitize(spec.FieldsMap["code"].Name),
		Sanitize(spec.FieldsMap["attempts"].Name), Sanitize(spec.FieldsMap["expires"].Name))
	from.From(Sanitize(spec.Name)).Where(from.Equal(Sanitize(filterField), filterValue))

	b := sqlbuilder.PostgreSQL.NewSelectBuilder()
	b.Select("row_to_json(t)")
	b.From(b.BuilderAs(from, "t"))

	sql, args := b.Build()
	return s.RawQuery(sql, args...)
}

func (s *Storage) IncrAttempts(spec *collections.Spec, filterField string, filterValue interface{}) error {
	b := sqlbuilder.PostgreSQL.NewUpdateBuilder()

	b.Update(Sanitize(spec.Name)).Set(b.Incr(Sanitize(spec.FieldsMap["attempts"].Name)))
	b.Where(b.Equal(Sanitize(filterField), filterValue))

	sql, args := b.Build()
	return s.RawExec(sql, args...)
}
