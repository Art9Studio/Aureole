package types

import (
	"aureole/internal/identity"
	"aureole/internal/plugins"
	"github.com/gofiber/fiber/v2"
)

type (
	AuthFunc func(fiber.Ctx) (*identity.Credential, fiber.Map, error)

	Authenticator interface {
		plugins.MetaDataGetter
		Login() AuthFunc
	}
)
