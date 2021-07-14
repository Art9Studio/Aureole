package types

import "aureole/internal/collections"

type PwResetData struct {
	// todo: try to use types
	Id      interface{}
	Email   interface{}
	Token   interface{}
	Expires interface{}
	Invalid interface{}
}

type PwReset interface {
	InsertReset(*collections.Spec, *PwResetData) (JSONCollResult, error)

	GetReset(*collections.Spec, []Filter) (JSONCollResult, error)

	InvalidateReset(*collections.Spec, []Filter) error
}
