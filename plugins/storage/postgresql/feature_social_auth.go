package postgresql

import (
	"aureole/internal/collections"
	"aureole/internal/plugins/storage/types"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
)

func (s *Storage) InsertSocialAuth(spec *collections.Spec, data *types.SocialAuthData) (types.JSONCollResult, error) {
	cols := []string{
		Sanitize(spec.FieldsMap["social_id"].Name),
		Sanitize(spec.FieldsMap["provider"].Name),
		Sanitize(spec.FieldsMap["user_data"].Name),
	}
	values := []interface{}{data.SocialId, data.Provider, data.UserData}

	if data.Email != nil {
		cols = append(cols, Sanitize(spec.FieldsMap["email"].Name))
		values = append(values, data.Email)
	}

	if data.UserId != nil {
		cols = append(cols, Sanitize(spec.FieldsMap["user_id"].Name))
		values = append(values, data.UserId)
	}

	for fieldName := range data.Additional {
		cols = append(cols, Sanitize(spec.FieldsMap[fieldName].Name))
		values = append(values, data.Additional[fieldName])
	}

	b := sqlbuilder.PostgreSQL.NewInsertBuilder()
	b.InsertInto(Sanitize(spec.Name))
	b.Cols(cols...)
	b.Values(values...)
	b.SQL(fmt.Sprintf(" returning %s", Sanitize(spec.Pk)))

	sql, args := b.Build()
	return s.RawQuery(sql, args...)
}

func (s *Storage) GetSocialAuth(spec *collections.Spec, filters []types.Filter) (types.JSONCollResult, error) {
	from := sqlbuilder.PostgreSQL.NewSelectBuilder()
	from.Select("*").From(Sanitize(spec.Name))

	for _, f := range filters {
		from.Where(from.Equal(Sanitize(f.Name), f.Value))
	}

	b := sqlbuilder.PostgreSQL.NewSelectBuilder()
	b.Select("row_to_json(t)")
	b.From(b.BuilderAs(from, "t"))
	sql, args := b.Build()

	return s.RawQuery(sql, args...)
}

func (s *Storage) IsSocialAuthExist(spec *collections.Spec, filters []types.Filter) (bool, error) {
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

func (s *Storage) LinkAccount(spec *collections.Spec, filters []types.Filter, userId interface{}) error {
	b := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	b.Update(Sanitize(spec.Name)).Set(b.Assign(Sanitize(spec.FieldsMap["user_id"].Name), userId))

	for _, f := range filters {
		b.Where(b.Equal(Sanitize(f.Name), f.Value))
	}

	b.SQL(fmt.Sprintf(" returning %s", Sanitize(spec.Pk)))
	sql, args := b.Build()

	return s.RawExec(sql, args...)
}
