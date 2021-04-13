package types

import (
	"aureole/internal"
	"github.com/gofiber/fiber/v2"
)

type Authorizer interface {
	internal.Initializer
	Authorize(*fiber.Ctx, map[string]interface{}) error
}
