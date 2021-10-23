package pwbased

import (
	"aureole/internal/collections"
)

var (
	passwordColType = &collections.CollectionType{
		Name:           "pwbased",
		IsAppendix:     true,
		ParentCollType: "identity",
	}
)
