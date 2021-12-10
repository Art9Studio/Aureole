package types

import (
	storageT "aureole/internal/plugins/storage/types"
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

func NewPayload(_ Authorizer, _ storageT.Storage, data map[string]interface{}) *Payload {
	p := &Payload{
		Id:       data["id"],
		SocialId: data["social_id"],
		Username: data["username"],
		Phone:    data["phone"],
		Email:    data["email"],
		UserData: data["user_data"],
	}

	/*if storage != nil {
		p.NativeQ = func(queryName string, args ...interface{}) string {
			queries := authorizer.GetNativeQueries()

			q, ok := queries[queryName]
			if !ok {
				return "--an error occurred during render--"
			}

			rawRes, err := storage.NativeQuery(q, args...)
			if err != nil {
				return "--an error occurred during render--"
			}

			res, err := json.Marshal(rawRes)
			if err != nil {
				return "--an error occurred during render--"
			}

			return string(res)
		}
	}*/

	return p
}
