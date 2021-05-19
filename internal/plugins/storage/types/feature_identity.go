package types

import (
	"aureole/internal/identity"
)

type IdentityData struct {
	Id         interface{}
	Username   interface{}
	Phone      interface{}
	Email      interface{}
	Additional map[string]interface{}
}

type Identity interface {
	// InsertIdentity inserts user entity in the user collection
	InsertIdentity(*identity.Identity, *IdentityData) (JSONCollResult, error)

	GetIdentity(*identity.Identity, string, interface{}) (JSONCollResult, error)

	IsIdentityExist(*identity.Identity, string, interface{}) (bool, error)
}
