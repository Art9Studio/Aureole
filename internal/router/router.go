package router

import (
	"aureole/internal/state/app"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"path"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type (
	IRouter interface {
		AddAppRoutes(appName string, routes []*Route)
		AddProjectRoutes(routes []*Route)
		GetAppRoutes() map[string][]*Route
		GetProjectRoutes() []*Route
	}

	Router struct {
		AppRoutes     map[string][]*Route
		ProjectRoutes []*Route
	}

	Route struct {
		Method  string
		Path    string
		Handler func(*fiber.Ctx) error
	}
)

const (
	MethodPOST = "POST"
	MethodGET  = "GET"
)

var (
	router *Router
	once   sync.Once
)

// CreateServer initializes router and creates routes for each application
func CreateServer(apps map[string]*app.App) (*fiber.App, error) {
	r := fiber.New(fiber.Config{DisableStartupMessage: true})
	r.Use(cors.New())
	r.Use(logger.New())
	v := r.Group("")

	for appName, routes := range router.AppRoutes {
		pathPrefix := apps[appName].PathPrefix
		appR := v.Group(pathPrefix)

		for _, route := range routes {
			appR.Add(route.Method, route.Path, route.Handler)
		}
	}

	for _, route := range router.ProjectRoutes {
		v.Add(route.Method, route.Path, route.Handler)
	}

	return r, nil
}

func GetRouter() IRouter {
	if router == nil {
		once.Do(
			func() {
				router = &Router{
					AppRoutes:     make(map[string][]*Route),
					ProjectRoutes: []*Route{},
				}
			})
	}
	return router
}

func (r *Router) AddAppRoutes(appName string, routes []*Route) {
	for i := range routes {
		routes[i].Path = path.Clean(routes[i].Path)
	}

	if existRoutes, ok := r.AppRoutes[appName]; ok {
		r.AppRoutes[appName] = append(existRoutes, routes...)
	} else {
		r.AppRoutes[appName] = routes
	}
}

func (r *Router) AddProjectRoutes(routes []*Route) {
	for i := range routes {
		routes[i].Path = path.Clean(routes[i].Path)
	}

	r.ProjectRoutes = append(r.ProjectRoutes, routes...)
}

func (r *Router) GetAppRoutes() map[string][]*Route {
	return r.AppRoutes
}

func (r *Router) GetProjectRoutes() []*Route {
	return r.ProjectRoutes
}

func SendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}
