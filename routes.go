package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gouth/authN"
)

// initRouter initializes router and creates routes for each application
func initRouter() (*fiber.App, error) {
	r := fiber.New()
	v := r.Group("v" + Project.APIVersion)

	for _, app := range Project.Apps {
		appR := v.Group(app.PathPrefix)
		for _, authNVariant := range app.AuthN {
			// todo: move it outside
			authController, err := authN.New(authNVariant.Type, &authNVariant)
			if err != nil {
				return nil, fmt.Errorf("router init error: %v", err)
			}

			for _, route := range authController.GetRoutes() {
				appR.Add(route.Method, route.Path, route.Handler)
			}
		}
	}

	return r, nil
}
