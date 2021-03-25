package types

import (
	"github.com/gofiber/fiber/v2"
)

type Authorizer interface {
	Init(string) error
	Authorize(*fiber.Ctx, *Context) error
}

type Context struct {
	UserId int
}
