package router

import (
	"aureole/internal/context/app"
	"aureole/internal/router/interface"
	"github.com/gofiber/fiber/v2"
	"path"
)

type TRouter struct {
	Routes map[string][]*_interface.Route
}

var Router TRouter

// CreateServer initializes router and creates routes for each application
func CreateServer(apps map[string]*app.App) (*fiber.App, error) {
	r := fiber.New()
	v := r.Group("")

	for appName, routes := range Router.Routes {
		pathPrefix := apps[appName].PathPrefix
		appR := v.Group(pathPrefix)

		for _, route := range routes {
			appR.Add(route.Method, route.Path, route.Handler)
		}
	}

	return r, nil
}

func Init() TRouter {
	Router = TRouter{
		Routes: make(map[string][]*_interface.Route),
	}
	return Router
}

func (r TRouter) Add(appName string, routes []*_interface.Route) {
	for i := range routes {
		routes[i].Path = path.Clean(routes[i].Path)
	}

	if existRoutes, ok := r.Routes[appName]; ok {
		r.Routes[appName] = append(existRoutes, routes...)
	} else {
		r.Routes[appName] = routes
	}
}
