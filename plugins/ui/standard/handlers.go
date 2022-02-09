package standard

import (
	"github.com/gofiber/fiber/v2"
)

func getRedirect(u *ui) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"redirect_url": u.conf.SuccessRedirect})
	}
}

func getJWTStorageKeys(u *ui) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"keys": map[string]string{
			"access":  u.conf.StorageJWTKeys.Access,
			"refresh": u.conf.StorageJWTKeys.Refresh,
		}})
	}
}
