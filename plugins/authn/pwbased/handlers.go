package pwbased

import (
	authzTypes "aureole/internal/plugins/authz/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"aureole/pkg/jsonpath"
	"errors"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func Login(context *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}

		if err := c.BodyParser(&authInput); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		usernamePath := context.conf.Login.FieldsMap["username"]
		username, err := getJsonField(authInput, usernamePath)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		passwordPath := context.conf.Login.FieldsMap["password"]
		password, err := getJsonField(authInput, passwordPath)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		// TODO: add a user existence check
		rawIdentity, err := context.storage.GetIdentity(context.identity, "username", username)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		identity, ok := rawIdentity.(map[string]interface{})
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": errors.New("unsupported identity type"),
			})
		}

		pw, err := context.storage.GetPassword(context.coll, "username", username)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		isMatch, err := context.pwHasher.ComparePw(password, pw.(string))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		if isMatch {
			// todo: add getUserId method
			collSpec := context.identity.Collection.Spec
			authzCtx := &authzTypes.Context{
				Username: username,
				UserId:   int(identity[collSpec.FieldsMap["id"]].(float64)),
			}
			return context.authorizer.Authorize(c, authzCtx)
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(&fiber.Map{
				"success": false,
				"message": "wrong password or username",
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
				"message": err.Error(),
			})
		}

		usernamePath := context.conf.Login.FieldsMap["username"]
		username, err := getJsonField(authInput, usernamePath)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		passwordPath := context.conf.Login.FieldsMap["password"]
		password, err := getJsonField(authInput, passwordPath)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		pwHash, err := context.pwHasher.HashPw(password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		// TODO: add a user existence check
		identityData := &storageTypes.IdentityData{Username: username}
		pwData := &storageTypes.PwBasedData{Password: pwHash}
		userId, err := context.storage.InsertPwBased(context.identity, identityData, context.coll, pwData)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		if context.conf.Register.IsLoginAfter {
			authzCtx := &authzTypes.Context{
				Username: username,
				UserId:   int(userId.(float64)),
			}
			return context.authorizer.Authorize(c, authzCtx)
		} else {
			return c.JSON(&fiber.Map{"id": userId})
		}
	}
}

func getJsonField(json interface{}, fieldPath string) (string, error) {
	rawData, err := jsonpath.GetJSONPath(fieldPath, json)
	if err != nil {
		return "", errors.New("field didn't passed")
	}

	data, ok := rawData.(string)
	if ok {
		if strings.TrimSpace(data) == "" {
			return "", errors.New("field can't be blank")
		}
	}

	return data, nil
}
