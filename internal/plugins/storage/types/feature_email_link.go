package types

import "aureole/internal/collections"

type EmailLinkData struct {
	Id      interface{}
	Email   interface{}
	Token   interface{}
	Expires interface{}
	Invalid interface{}
}

type EmailLink interface {
	InsertEmailLink(*collections.Spec, *EmailLinkData) (JSONCollResult, error)

	GetEmailLink(*collections.Spec, string, interface{}) (JSONCollResult, error)

	InvalidateEmailLink(*collections.Spec, string, interface{}) error
}
