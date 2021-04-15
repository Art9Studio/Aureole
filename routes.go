package main

import (
	"github.com/gofiber/fiber/v2"
)

// initRouter initializes router and creates routes for each application
func initRouter() (*fiber.App, error) {
	r := fiber.New()
	v := r.Group("")

	for _, app := range Project.Apps {
		appR := v.Group(app.PathPrefix)
		for _, route := range Project.Routes {
			appR.Add(route.Method, route.Path, route.Handler)
		}

	}

	return r, nil
}
