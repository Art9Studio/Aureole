package types

import (
	"github.com/gofiber/fiber/v2"
)

type Authenticator interface {
	Initialize(string) error
	GetRoutes() []Route
}

type Route struct {
	Method  string
	Path    string
	Handler func(*fiber.Ctx) error
}
