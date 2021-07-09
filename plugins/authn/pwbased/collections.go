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

	resetColType = &collections.CollectionType{
		Name:       "password_reset",
		IsAppendix: false,
	}

	verifColType = &collections.CollectionType{
		Name:       "email_verification",
		IsAppendix: false,
	}
)
