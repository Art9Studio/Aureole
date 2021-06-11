package types

import (
	"aureole/internal/collections"
)

type PhoneVerificationData struct {
	// todo: try to use types
	Id       interface{}
	Phone    interface{}
	Code     interface{}
	Attempts interface{}
	Expires  interface{}
}

type PhoneVerification interface {
	InsertVerification(*collections.Spec, *PhoneVerificationData) (JSONCollResult, error)

	GetVerification(*collections.Spec, string, interface{}) (JSONCollResult, error)

	IncrAttempts(*collections.Spec, string, interface{}) error
}
