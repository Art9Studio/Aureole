package main

import (
	"github.com/gofiber/fiber/v2"
)

// initRouter initializes router and creates routes for each application
func initRouter() (*fiber.App, error) {
	r := fiber.New()
	//v := r.Group("v" + Project.APIVersion)
	v := r.Group("")

	for _, app := range Project.Apps {
		appR := v.Group(app.PathPrefix)
		for _, controller := range app.AuthnControllers {
			for _, route := range controller.GetRoutes() {
				appR.Add(route.Method, route.Path, route.Handler)
			}
		}
	}

	return r, nil
}
