package types

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

type Type int

const (
	PasswordBased Type = iota
	Passwordless
)

var typeNames = [...]string{"password_based", "passwordless"}

func ToAuthnType(authType string) (Type, error) {
	for i, name := range typeNames {
		if name == authType {
			return Type(i), nil
		}
	}

	return 0, fmt.Errorf("authenticate type '%s' is not declared", authType)
}

func (t Type) String() string {
	return typeNames[t]
}

type Controller interface {
	GetRoutes() []Route
}

type Route struct {
	Method  string
	Path    string
	Handler func(*fiber.Ctx) error
}
