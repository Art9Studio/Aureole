package pwbased

import (
	storageTypes "aureole/internal/plugins/storage/types"
	"aureole/jsonpath"
	"aureole/jwt"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func Login(context *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}

		if err := c.BodyParser(&authInput); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		identityPath := context.Conf.Login.FieldsMap["identity"]
		IIdentity, err := jsonpath.GetJSONPath(identityPath, authInput)
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

		passwordPath := context.Conf.Login.FieldsMap["password"]
		IPassword, err := jsonpath.GetJSONPath(passwordPath, authInput)
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

func Register(context *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}

		if err := c.BodyParser(&authInput); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		identityPath := context.Conf.Login.FieldsMap["identity"]
		IIdentity, err := jsonpath.GetJSONPath(identityPath, authInput)
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

		passwordPath := context.Conf.Login.FieldsMap["password"]
		IPassword, err := jsonpath.GetJSONPath(passwordPath, authInput)
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

		pwHash, err := context.PwHasher.HashPw(password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		// TODO: add a user existence check
		identityStorage := context.Storage
		insertData := storageTypes.InsertIdentityData{
			Identity:    identity,
			UserConfirm: pwHash,
		}
		res, err := identityStorage.InsertIdentity(context.IdentityColl.Spec, insertData)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		if context.Conf.Register.IsLoginAfter {
			token := jwt.IssueToken()
			return c.JSON(&fiber.Map{"token": token})
		} else {
			return c.JSON(&fiber.Map{"id": res})
		}
	}
}
