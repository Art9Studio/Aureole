package types

import "github.com/gofiber/fiber/v2"

type Authorizer interface {
	Authorize(*fiber.Ctx, map[string]interface{}) error
}
