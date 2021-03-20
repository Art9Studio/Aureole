package types

import (
	"github.com/gofiber/fiber/v2"
)

type Controller interface {
	GetRoutes() []Route
	// todo: add method successAuthz, which works with authorization
}

type Route struct {
	Method  string
	Path    string
	Handler func(*fiber.Ctx) error
}
