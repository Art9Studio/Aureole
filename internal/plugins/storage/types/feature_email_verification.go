package types

import "aureole/internal/collections"

type EmailVerifData struct {
	Id      interface{}
	Email   interface{}
	Token   interface{}
	Expires interface{}
	Invalid interface{}
}

type EmailVerification interface {
	InsertEmailVerif(*collections.Spec, *EmailVerifData) (JSONCollResult, error)

	GetEmailVerif(*collections.Spec, []Filter) (JSONCollResult, error)

	InvalidateEmailVerif(*collections.Spec, []Filter) error
}
