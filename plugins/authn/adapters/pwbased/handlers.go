package pwbased

import (
	"aureole/jsonpath"
	"aureole/jwt"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func Auth(context *Ctx) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}

		if err := c.BodyParser(&authInput); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		IIdentity, err := jsonpath.GetJSONPath(context.Identity, authInput)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": "identity didn't passed",
			})
		}

		identity, ok := IIdentity.(string)
		if ok {
			if strings.TrimSpace(identity) == "" {
				return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
					"success": false,
					"message": "identity can't be blank",
				})
			}
		}

		IPassword, err := jsonpath.GetJSONPath(context.Password, authInput)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": "password didn't passed",
			})
		}

		password, ok := IPassword.(string)
		if ok {
			if strings.TrimSpace(password) == "" {
				return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
					"success": false,
					"message": "password can't be blank",
				})
			}
		}

		// TODO: add a user existence check
		identityStorage := context.Storage
		pw, err := identityStorage.GetPasswordByIdentity(context.IdentityColl.Spec, identity)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		isMatch, err := context.PwHasher.ComparePw(password, pw.(string))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		if isMatch {
			token := jwt.IssueToken()
			return c.JSON(&fiber.Map{"token": token})
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}
	}
}
