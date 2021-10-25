package types

import (
	"aureole/internal/collections"

	"github.com/gofiber/fiber/v2"
)

type Authorizer interface {
	Init(string) error
	GetNativeQueries() map[string]string
	Authorize(*fiber.Ctx, *Payload) error
}

type Payload struct {
	Id         interface{}
	SocialId   interface{}
	Username   interface{}
	Phone      interface{}
	Email      interface{}
	UserData   interface{}
	Additional map[string]interface{}
	NativeQ    func(queryName string, args ...interface{}) string
}

func NewPayload(identity map[string]interface{}, collMap map[string]collections.FieldSpec) *Payload {
	p := &Payload{
		Id:         identity[collMap["id"].Name],
		Username:   identity[collMap["username"].Name],
		Phone:      identity[collMap["phone"].Name],
		Email:      identity[collMap["email"].Name],
		Additional: map[string]interface{}{},
	}

	for fieldName, fieldVal := range identity {
		if fieldName != collMap["id"].Name &&
			fieldName != collMap["username"].Name &&
			fieldName != collMap["email"].Name &&
			fieldName != collMap["phone"].Name {
			p.Additional[fieldName] = fieldVal
		}
	}

	return p
}
