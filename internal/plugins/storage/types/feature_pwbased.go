package types

import (
	"aureole/internal/collections"
	"aureole/internal/identity"
)

type PwBasedData struct {
	Password     interface{}
	PasswordHash string
}

type PwBased interface {
	InsertPwBased(*identity.Identity, *collections.Collection, *IdentityData, *PwBasedData) (JSONCollResult, error)

	GetPassword(*collections.Collection, string, interface{}) (JSONCollResult, error)
}
