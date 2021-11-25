package types

import (
	"aureole/internal/plugins"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

type Authorizer interface {
	plugins.MetaDataGetter
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
	// NativeQ    func(queryName string, args ...interface{}) string
}

func NewPayload(data map[string]interface{}) (*Payload, error) {
	p := &Payload{}
	if err := mapstructure.Decode(data, p); err != nil {
		return nil, err
	}
	return p, nil
}
