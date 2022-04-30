package plugins

import (
	"github.com/go-openapi/spec"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
)

type (
	Issuer interface {
		MetaDataGetter
		GetResponseData() (*spec.Responses, spec.Definitions)
		GetNativeQueries() map[string]string
		Authorize(*fiber.Ctx, *Payload) error
	}

	Payload struct {
		ID            interface{}            `mapstructure:"id,omitempty"`
		Username      *string                `mapstructure:"username,omitempty"`
		Phone         *string                `mapstructure:"phone,omitempty"`
		Email         *string                `mapstructure:"email,omitempty"`
		EmailVerified bool                   `mapstructure:"email_verified"`
		PhoneVerified bool                   `mapstructure:"phone_verified"`
		Additional    map[string]interface{} `mapstructure:"additional,omitempty"`
		// NativeQ    func(queryName string, args ...interface{}) string
	}
)

func NewPayload(data map[string]interface{}) (*Payload, error) {
	p := &Payload{}
	if err := mapstructure.Decode(data, p); err != nil {
		return nil, err
	}
	return p, nil
}
