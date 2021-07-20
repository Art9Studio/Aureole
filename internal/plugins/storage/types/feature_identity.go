package types

import (
	"aureole/internal/collections"
	"aureole/internal/identity"
)

type IdentityData struct {
	Id         interface{}
	Username   interface{}
	Phone      interface{}
	Email      interface{}
	Additional map[string]interface{}
}

func NewIdentityData(rawData JSONCollResult, specs map[string]collections.FieldSpec) *IdentityData {
	data := rawData.(map[string]interface{})

	socAuth := &IdentityData{
		Id:         data[specs["id"].Name],
		Username:   data[specs["username"].Name],
		Phone:      data[specs["phone"].Name],
		Email:      data[specs["email"].Name],
		Additional: map[string]interface{}{},
	}

	for fieldName, fieldVal := range data {
		if fieldName != specs["id"].Name &&
			fieldName != specs["username"].Name &&
			fieldName != specs["phone"].Name &&
			fieldName != specs["email"].Name {
			socAuth.Additional[fieldName] = fieldVal
		}
	}

	return socAuth
}

type Identity interface {
	// InsertIdentity inserts user entity in the user collection
	InsertIdentity(*identity.Identity, *IdentityData) (JSONCollResult, error)

	GetIdentity(*identity.Identity, []Filter) (JSONCollResult, error)

	IsIdentityExist(*identity.Identity, []Filter) (bool, error)

	SetEmailVerified(*collections.Spec, []Filter) error

	SetPhoneVerified(*collections.Spec, []Filter) error
}
