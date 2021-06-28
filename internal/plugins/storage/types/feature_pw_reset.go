package types

import "aureole/internal/collections"

type PwResetData struct {
	// todo: try to use types
	Id      interface{}
	Email   interface{}
	Token   interface{}
	Expires interface{}
}

type PwReset interface {
	InsertReset(*collections.Spec, *PwResetData) (JSONCollResult, error)

	GetReset(*collections.Spec, string, interface{}) (JSONCollResult, error)
}
