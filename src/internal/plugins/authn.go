package plugins

import (
	"github.com/gofiber/fiber/v2"
)

type (
	Authenticator interface {
		MetaDataGetter
		OpenAPISpecGetter
		LoginWrapper() AuthNLoginFunc
	}

	AuthNResult struct {
		Cred       *Credential
		Identity   *Identity
		Provider   string
		Additional map[string]interface{}
		ErrorData  map[string]interface{}
	}

	AuthNLoginFunc func(fiber.Ctx) (*AuthNResult, error)
)
