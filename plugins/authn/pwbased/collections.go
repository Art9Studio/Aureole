package pwbased

import (
	"aureole/internal/collections"
	"aureole/internal/plugins/authn"
)

var passwordColType = &collections.CollectionType{
	Name:           "pwbased",
	IsAppendix:     true,
	ParentCollType: "identity",
}

func registerCollectionTypes() error {
	return authn.Repository.PluginApi.RegisterCollectionType(passwordColType)
}
