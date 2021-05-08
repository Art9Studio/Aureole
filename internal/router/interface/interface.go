package _interface

import (
	"github.com/gofiber/fiber/v2"
)

type IRouter interface {
	Add(appName string, routes []*Route)
}

type Route struct {
	Method  string
	Path    string
	Handler func(*fiber.Ctx) error
}
