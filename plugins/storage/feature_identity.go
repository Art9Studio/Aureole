package storage

import "aureole/collections"

type InsertIdentityData struct {
	Identity    interface{}
	UserConfirm interface{}
}

type Application interface {
	// IsCollExists checks whether the given collection exists
	IsCollExists(collections.Specification) (bool, error)

	// CreateUserColl creates user collection with traits passed by UserCollectionConfig
	CreateIdentityColl(collections.Specification) error

	// InsertUser inserts user entity in the user collection
	InsertIdentity(collections.Specification, InsertIdentityData) (JSONCollResult, error)

	GetPasswordByIdentity(collections.Specification, interface{}) (JSONCollResult, error)
}

func NewInsertIdentityData(userUnique interface{}, userConfirm interface{}) InsertIdentityData {
	return InsertIdentityData{Identity: userUnique, UserConfirm: userConfirm}
}
