package types

import (
	"aureole/internal/identity"
)

type IdentityData struct {
	Id       interface{}
	Username interface{}
	Phone    interface{}
	Email    interface{}
}

type Identity interface {
	// CreateIdentityColl creates user collection with traits passed by UserCollectionConfig
	CreateIdentityColl(*identity.Identity) error

	// InsertIdentity inserts user entity in the user collection
	InsertIdentity(*identity.Identity, *IdentityData) (JSONCollResult, error)

	GetIdentity(*identity.Identity, string, interface{}) (JSONCollResult, error)
}
