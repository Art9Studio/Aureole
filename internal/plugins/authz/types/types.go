package types

import (
	"aureole/internal/collections"

	"github.com/gofiber/fiber/v2"
)

type Authorizer interface {
	Init(string) error
	GetNativeQueries() map[string]string
	Authorize(*fiber.Ctx, *Context) error
}

type Context struct {
	Id         interface{}
	Username   interface{}
	Phone      interface{}
	Email      interface{}
	Additional map[string]interface{}
	NativeQ    func(queryName string, args ...interface{}) string
}

func NewContext(identity map[string]interface{}, fieldsMap map[string]collections.FieldSpec) *Context {
	context := &Context{
		Id:         identity[fieldsMap["id"].Name],
		Username:   identity[fieldsMap["username"].Name],
		Phone:      identity[fieldsMap["phone"].Name],
		Email:      identity[fieldsMap["email"].Name],
		Additional: map[string]interface{}{},
	}

	for fieldName, fieldVal := range identity {
		if fieldName != fieldsMap["id"].Name &&
			fieldName != fieldsMap["username"].Name &&
			fieldName != fieldsMap["email"].Name &&
			fieldName != fieldsMap["phone"].Name {
			context.Additional[fieldName] = fieldVal
		}
	}

	return context
}
