package pwbased

import (
	"aureole/internal/identity"
	"errors"
	"github.com/gofiber/fiber/v2"
	"net/url"
)

func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(&fiber.Map{
		"success": false,
		"message": message,
	})
}

func getCredential(i *identity.Identity) (*identity.Credential, error) {
	if i.Username != "nil" {
		return &identity.Credential{
			Name:  "username",
			Value: i.Username,
		}, nil
	}

	if i.Email != "nil" {
		return &identity.Credential{
			Name:  "email",
			Value: i.Email,
		}, nil
	}

	if i.Phone != "nil" {
		return &identity.Credential{
			Name:  "phone",
			Value: i.Phone,
		}, nil
	}

	return nil, errors.New("credential not found")
}

func attachToken(u *url.URL, token string) string {
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	return u.String()
}
