package types

import (
	"aureole/internal/plugins/core"
	"github.com/gofiber/fiber/v2"
)

type Authorizer interface {
	Initialize() error
	Authorize(*fiber.Ctx, map[string]interface{}) error
	GetRoutes() []*core.Route
}
