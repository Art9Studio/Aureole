package identity

import (
	"aureole/internal/collections"
)

var identColType = &collections.CollectionType{
	Name: "identity",
}

func RegisterCollectionTypes() error {
	collections.Repository.Register(identColType)
	return nil
}
