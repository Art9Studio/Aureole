package core

import (
	fiberSwagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/getkin/kin-openapi/openapi3"
	"os"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type (
	Route struct {
		Method    string
		Path      string
		Operation *openapi3.Operation
		Handler   func(c *fiber.Ctx) error
	}

	PathsGetter interface {
		GetPaths() []*Route
	}

	ErrorMessage struct {
		Error string
	}

	router struct {
		appRoutes     map[string][]*Route
		projectRoutes []*Route
	}
)

func RunServer(p *project, r *router) error {
	var port string
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "3000"
	}
	return createServer(p, r).Listen(":" + port)
}

// createServer initializes router and creates routes for each application
func createServer(p *project, r *router) *fiber.App {
	fiberApp := fiber.New(fiber.Config{DisableStartupMessage: true})
	fiberApp.Use(cors.New())
	fiberApp.Use(logger.New())

	for appName, routes := range r.appRoutes {
		pathPrefix := p.apps[appName].pathPrefix
		appGroup := fiberApp.Group(pathPrefix)

		for _, route := range routes {
			appGroup.Add(route.Method, route.Path, route.Handler)
		}
	}

	for _, route := range r.projectRoutes {
		fiberApp.Add(route.Method, route.Path, route.Handler)
	}

	fiberApp.Get("/openapi/*", fiberSwagger.HandlerDefault)
	return fiberApp
}

func CreateRouter() *router {
	return &router{
		appRoutes:     make(map[string][]*Route),
		projectRoutes: []*Route{},
	}
}

func (r *router) addAppRoutes(appName string, routes []*Route) {
	for i := range routes {
		routes[i].Path = path.Clean(routes[i].Path)
	}

	if existRoutes, ok := r.appRoutes[appName]; ok {
		r.appRoutes[appName] = append(existRoutes, routes...)
	} else {
		r.appRoutes[appName] = routes
	}
}

func (r *router) addProjectRoutes(routes []*Route) {
	for i := range routes {
		routes[i].Path = path.Clean(routes[i].Path)
	}

	r.projectRoutes = append(r.projectRoutes, routes...)
}

func (r *router) getAppRoutes() map[string][]*Route {
	return r.appRoutes
}

func (r *router) getProjectRoutes() []*Route {
	return r.projectRoutes
}

func SendError(c *fiber.Ctx, statusCode int, errorMessage string) error {
	return c.Status(statusCode).JSON(ErrorMessage{Error: errorMessage})
}
