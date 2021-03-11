package authN

import (
	"github.com/gofiber/fiber/v2"
	"gouth/authN/types"
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

func New(authType types.Type, conf *config.AuthNConfig) (Controller, error) {
	adapter, err := GetAdapter(authType.String())
	if err != nil {
		return nil, err
	}

	return adapter.GetAuthNController(conf.Path, &conf.Config)
}
