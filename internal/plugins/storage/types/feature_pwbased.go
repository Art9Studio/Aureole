package types

import (
	"aureole/internal/collections"
	"aureole/internal/identity"
)

type PwBasedData struct {
	Password interface{}
}

type PwBased interface {
	CreatePwBasedColl(*collections.Collection) error

	InsertPwBased(*identity.Identity, *IdentityData, *collections.Collection, *PwBasedData) (JSONCollResult, error)

	GetPassword(*collections.Collection, string, interface{}) (JSONCollResult, error)
}
