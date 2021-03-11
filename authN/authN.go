package authN

import (
	"github.com/gofiber/fiber/v2"
	"gouth/config"
)

type Controller interface {
	GetRoutes() []Route
}

type Route struct {
	Method  string
	Path    string
	Handler func(*fiber.Ctx) error
}

func New(conf *config.AuthNConfig) (Controller, error) {
	adapter, err := GetAdapter(conf.Type.String())
	if err != nil {
		return nil, err
	}

	return adapter.GetAuthNController(conf.PathPrefix, &conf.Config, projectCtx)
}
