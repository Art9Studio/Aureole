package session

import (
	"aureole/internal/collections"
)

var sessionColType = &collections.CollectionType{
	Name: "session",
}

func registerCollectionTypes() error {
	collections.Repository.Register(sessionColType)
	return nil
}
