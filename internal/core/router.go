package core

import (
	fiberSwagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"net"
	_ "net/http/pprof"
	"os"
	"path"
	"sync"
	"syscall"
)

var (
	r    *router
	once sync.Once
)

type (
	Route struct {
		Method  string
		Path    string
		Handler func(c *fiber.Ctx) error
	}

	ErrorMessage struct {
		Error string
	}

	router struct {
		appRoutes     map[string][]*Route
		projectRoutes []*Route
	}
)

func RunServer(ln net.Listener) error {
	return createServer().Listener(ln)
}

// createServer initializes router and creates routes for each application
func createServer() *fiber.App {
	fiberApp := fiber.New(fiber.Config{DisableStartupMessage: true})
	fiberApp.Use(cors.New())
	fiberApp.Use(logger.New())
	fiberApp.Use(pprof.New())

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

	fiberApp.Get("/swagger/*", fiberSwagger.HandlerDefault)
	fiberApp.Get("/reload", reload)
	return fiberApp
}

func reload(c *fiber.Ctx) error {
	// todo: make this route secure
	err := syscall.Kill(os.Getppid(), syscall.SIGUSR2)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusOK)
}

func getRouter() *router {
	once.Do(
		func() {
			r = &router{
				appRoutes:     make(map[string][]*Route),
				projectRoutes: []*Route{},
			}
		})
	return r
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
