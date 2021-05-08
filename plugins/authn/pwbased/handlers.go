package pwbased

import (
	storageTypes "aureole/internal/plugins/storage/types"
	"aureole/pkg/jsonpath"
	"errors"
	"fmt"
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

		usernamePath := context.conf.Login.FieldsMap["username"]
		username, err := getJsonField(authInput, usernamePath)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": fmt.Sprintf("username: %v", err),
			})
		}

		passwordPath := context.conf.Login.FieldsMap["password"]
		password, err := getJsonField(authInput, passwordPath)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": fmt.Sprintf("password: %v", err),
			})
		}

		// TODO: add a user existence check
		rawIdentity, err := context.storage.GetIdentity(context.identity, "username", username)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err,
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
			collSpec := context.identity.Collection.Spec
			return context.authorizer.Authorize(c, map[string]interface{}{"user_id": identity[collSpec.FieldsMap["id"]]})
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

		usernamePath := context.conf.Login.FieldsMap["username"]
		username, err := getJsonField(authInput, usernamePath)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": fmt.Sprintf("username: %v", err),
			})
		}

		passwordPath := context.conf.Login.FieldsMap["password"]
		password, err := getJsonField(authInput, passwordPath)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"success": false,
				"message": fmt.Sprintf("password: %v", err),
			})
		}

		pwHash, err := context.pwHasher.HashPw(password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
				"success": false,
				"message": err,
			})
		}

		// TODO: add a user existence check
		identityData := &storageTypes.IdentityData{Username: username}
		pwData := &storageTypes.PwBasedData{Password: pwHash}
		userId, err := context.storage.InsertPwBased(context.identity, identityData, context.coll, pwData)
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
