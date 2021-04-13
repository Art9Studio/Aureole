package types

import (
	"aureole/internal"
	"github.com/gofiber/fiber/v2"
)

type Authenticator interface {
	internal.Initializer
	GetRoutes() []Route
}

type Route struct {
	Method  string
	Path    string
	Handler func(*fiber.Ctx) error
}
