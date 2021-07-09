package types

import (
	"aureole/internal/collections"
)

type PhoneVerificationData struct {
	// todo: try to use types
	Id       interface{}
	Phone    interface{}
	Otp      interface{}
	Attempts interface{}
	Expires  interface{}
	Invalid  interface{}
}

type PhoneVerification interface {
	InsertVerification(*collections.Spec, *PhoneVerificationData) (JSONCollResult, error)

	GetVerification(*collections.Spec, string, interface{}) (JSONCollResult, error)

	IncrAttempts(*collections.Spec, string, interface{}) error

	InvalidateVerification(*collections.Spec, string, interface{}) error
}
