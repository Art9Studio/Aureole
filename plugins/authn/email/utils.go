package email

import (
	"aureole/internal/identity"
	"github.com/gofiber/fiber/v2"
	"net/url"
)

func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(&fiber.Map{
		"success": false,
		"message": message,
	})
}

func isCredential(trait *identity.Trait) bool {
	return trait.IsCredential && trait.IsUnique
}

func initMagicLink(u *url.URL, token string) string {
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	return u.String()
}
