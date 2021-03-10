package authN

import (
	"github.com/gofiber/fiber/v2"
	"gouth/config"
)

type Controller interface {
	GetRoutes() []Route
}

type Route struct {
	Path    string
	Handler func(*fiber.Ctx) error
}

type Type int

const (
	PasswordBased Type = iota
	Passwordless
)

func (t Type) String() string {
	return [...]string{"password_based", "passwordless"}[t]
}

func New(authType Type, rawConf *config.RawConfig) (Controller, error) {
	adapter, err := GetAdapter(authType.String())
	if err != nil {
		return nil, err
	}

	return adapter.GetAuthNController(rawConf)
}
