package types

import (
	"github.com/gofiber/fiber/v2"
)

type Authenticator interface {
	GetRoutes() []Route
}

type Route struct {
	Method  string
	Path    string
	Handler func(*fiber.Ctx) error
}
