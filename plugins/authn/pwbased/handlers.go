package pwbased

import (
	storageTypes "aureole/internal/plugins/storage/types"
	"aureole/jsonpath"
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

		identityPath := context.conf.Login.FieldsMap["identity"]
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

		passwordPath := context.conf.Login.FieldsMap["password"]
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
		identityStorage := context.storage
		pw, err := identityStorage.GetPasswordByIdentity(context.identityColl.Spec, identity)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		isMatch, err := context.pwHasher.ComparePw(password, pw.(string))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		if isMatch {
			// todo: add getUserId method
			return context.authorizer.Authorize(c, map[string]interface{}{"user_id": 0})
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

		identityPath := context.conf.Login.FieldsMap["identity"]
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

		passwordPath := context.conf.Login.FieldsMap["password"]
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

		pwHash, err := context.pwHasher.HashPw(password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		// TODO: add a user existence check
		identityStorage := context.storage
		insertData := storageTypes.InsertIdentityData{
			Identity:    identity,
			UserConfirm: pwHash,
		}
		userId, err := identityStorage.InsertIdentity(context.identityColl.Spec, insertData)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		if context.conf.Register.IsLoginAfter {
			return context.authorizer.Authorize(c, map[string]interface{}{"user_id": userId})
		} else {
			return c.JSON(&fiber.Map{"id": userId})
		}
	}
}
