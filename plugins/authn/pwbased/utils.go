package pwbased

import (
	"github.com/gofiber/fiber/v2"
	"reflect"
)

func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(&fiber.Map{
		"success": false,
		"message": message,
	})
}

func getValueOrDefault(value, defaultValue interface{}) interface{} {
	if !isZeroVal(value) {
		return value
	} else if !isZeroVal(defaultValue) {
		return defaultValue
	} else {
		return nil
	}
}

func isZeroVal(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
