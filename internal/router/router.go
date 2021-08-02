package router

import (
	"aureole/internal/context/app"
	_interface "aureole/internal/router/interface"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type TRouter struct {
	AppRoutes     map[string][]*_interface.Route
	ProjectRoutes []*_interface.Route
}

var Router *TRouter

// CreateServer initializes router and creates routes for each application
func CreateServer(apps map[string]*app.App) (*fiber.App, error) {
	r := fiber.New(fiber.Config{AppName: "Aureole"})
	r.Use(cors.New())
	v := r.Group("")

	for appName, routes := range Router.AppRoutes {
		pathPrefix := apps[appName].PathPrefix
		appR := v.Group(pathPrefix)

		for _, route := range routes {
			appR.Add(route.Method, route.Path, route.Handler)
		}
	}

	for _, route := range Router.ProjectRoutes {
		v.Add(route.Method, route.Path, route.Handler)
	}

	return r, nil
}

func Init() *TRouter {
	Router = &TRouter{
		AppRoutes:     make(map[string][]*_interface.Route),
		ProjectRoutes: []*_interface.Route{},
	}
	return Router
}

func (r *TRouter) AddAppRoutes(appName string, routes []*_interface.Route) {
	for i := range routes {
		routes[i].Path = path.Clean(routes[i].Path)
	}

	if existRoutes, ok := r.AppRoutes[appName]; ok {
		r.AppRoutes[appName] = append(existRoutes, routes...)
	} else {
		r.AppRoutes[appName] = routes
	}
}

func (r *TRouter) AddProjectRoutes(routes []*_interface.Route) {
	for i := range routes {
		routes[i].Path = path.Clean(routes[i].Path)
	}

	r.ProjectRoutes = append(r.ProjectRoutes, routes...)
}

func (r *TRouter) GetAppRoutes() map[string][]*_interface.Route {
	return r.AppRoutes
}

func (r *TRouter) GetProjectRoutes() []*_interface.Route {
	return r.ProjectRoutes
}
