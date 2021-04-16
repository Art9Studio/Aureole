package router

import (
	contextTypes "aureole/context/types"
	"github.com/gofiber/fiber/v2"
)

type Route struct {
	Method  string
	Path    string
	Handler func(*fiber.Ctx) error
}

var Routes []*Route

// initRouter initializes router and creates routes for each application
func InitRouter(ctx *contextTypes.ProjectCtx) (*fiber.App, error) {
	r := fiber.New()
	v := r.Group("")

	for _, app := range ctx.Apps {
		appR := v.Group(app.PathPrefix)
		for _, route := range Routes {
			appR.Add(route.Method, route.Path, route.Handler)
		}

	}

	return r, nil
}
