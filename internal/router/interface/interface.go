package _interface

import "github.com/gofiber/fiber/v2"

type Router interface {
	AddAppRoutes(appName string, routes []*Route)
	AddProjectRoutes(routes []*Route)
	GetAppRoutes() map[string][]*Route
	GetProjectRoutes() []*Route
}

type Route struct {
	Method  string
	Path    string
	Handler func(*fiber.Ctx) error
}
