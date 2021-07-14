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

	GetVerification(*collections.Spec, []Filter) (JSONCollResult, error)

	IncrAttempts(*collections.Spec, []Filter) error

	InvalidateVerification(*collections.Spec, []Filter) error
}
