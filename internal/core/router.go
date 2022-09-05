package core

import (
	"os"
	"path"

	fiberSwagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

const AuthPipelinePath = "/auth"

type (
	Route struct {
		Method        string
		Path          string
		OAS3Operation *openapi3.Operation
		Handler       func(c *fiber.Ctx) error
	}

	ExtendedRoute struct {
		Route
		Metadata
	}

	RoutesGetter interface {
		GetCustomAppRoutes() []*Route
	}

	ErrorMessage struct {
		Error string
	}

	router struct {
		appRoutes     map[string][]*ExtendedRoute
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

	for _, routes := range r.getAppRoutes() {
		for _, route := range routes {
			fiberApp.Add(route.Method, route.Path, route.Handler)
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
		appRoutes:     make(map[string][]*ExtendedRoute),
		projectRoutes: []*Route{},
	}
}

func (r *router) addAppRoutes(appName string, routes []*ExtendedRoute) {
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

func (r *router) getAppRoutes() map[string][]*ExtendedRoute {
	return r.appRoutes
}

func (r *router) getProjectRoutes() []*Route {
	return r.projectRoutes
}

func SendToken(c *fiber.Ctx, token string) error {
	return c.JSON(&fiber.Map{"token": token})
}

func GetOAuthPathPostfix() string {
	return "/oauth"
}
