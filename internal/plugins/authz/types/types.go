package types

import (
	"github.com/gofiber/fiber/v2"
)

type Authorizer interface {
	Init() error
	Authorize(*fiber.Ctx, map[string]interface{}) error
}
