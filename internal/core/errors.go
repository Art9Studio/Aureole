package core

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

var ErrNoUser = errors.New("user not found")
var ErrDB = errors.New("error from database")

func WrapErrDB(errText string) error {
	return fmt.Errorf("%w: %s", ErrDB, errText)
}

func SendError(c *fiber.Ctx, statusCode int, errorMessage string) error {
	return c.Status(statusCode).JSON(ErrorMessage{Error: errorMessage})
}
