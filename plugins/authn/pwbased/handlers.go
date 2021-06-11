package pwbased

import (
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func Login(context *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		identityData := &storageT.IdentityData{}
		getLoginData(context, authInput, context.conf.Login.FieldsMap, identityData)

		credName, credVal, statusCode, err := getCredField(context, identityData)
		if err != nil {
			return sendError(c, statusCode, err.Error())
		}

		pwData := &storageT.PwBasedData{}
		if statusCode, err := getPwData(authInput, context.conf.Login.FieldsMap, pwData); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		exist, err := context.storage.IsIdentityExist(context.identity, credName, credVal)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if !exist {
			return sendError(c, fiber.StatusUnauthorized, "user doesn't exist")
		}

		rawIdentity, err := context.storage.GetIdentity(context.identity, credName, credVal)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		i, ok := rawIdentity.(map[string]interface{})
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "cannot get identity from database")
		}

		pw, err := context.storage.GetPassword(context.coll, credName, credVal)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		isMatch, err := context.pwHasher.ComparePw(pwData.Password.(string), pw.(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if isMatch {
			collSpec := context.identity.Collection.Spec
			authzCtx := authzT.NewContext(i, collSpec.FieldsMap)
			// todo: refactor this
			authzCtx.NativeQ = func(queryName string, args ...interface{}) string {
				queries := context.authorizer.GetNativeQueries()

				q, ok := queries[queryName]
				if !ok {
					return "--an error occurred during render--"
				}

				rawRes, err := context.storage.NativeQuery(q, args...)
				if err != nil {
					return "--an error occurred during render--"
				}

				res, err := json.Marshal(rawRes)
				if err != nil {
					return "--an error occurred during render--"
				}

				return string(res)
			}
			return context.authorizer.Authorize(c, authzCtx)
		} else {
			return sendError(c, fiber.StatusUnauthorized, fmt.Sprintf("wrong password or %s", credName))
		}
	}
}

func Register(context *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		identityData := &storageT.IdentityData{Additional: map[string]interface{}{}}
		if statusCode, err := getRegisterData(context, authInput, context.conf.Register.FieldsMap, identityData); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		pwData := &storageT.PwBasedData{}
		if statusCode, err := getPwData(authInput, context.conf.Register.FieldsMap, pwData); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		pwHash, err := context.pwHasher.HashPw(pwData.Password.(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		pwData.PasswordHash = pwHash

		credName, credVal, statusCode, err := getCredField(context, identityData)
		if err != nil {
			return sendError(c, statusCode, err.Error())
		}

		exist, err := context.storage.IsIdentityExist(context.identity, credName, credVal)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if exist {
			return sendError(c, fiber.StatusBadRequest, "user already exist")
		}

		id, err := context.storage.InsertPwBased(context.identity, context.coll, identityData, pwData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if context.conf.Register.IsLoginAfter {
			authzCtx := authzT.Context{
				Id:         identityData.Id,
				Username:   identityData.Username,
				Phone:      identityData.Phone,
				Email:      identityData.Email,
				Additional: identityData.Additional,
			}
			// todo: refactor this
			authzCtx.NativeQ = func(queryName string, args ...interface{}) string {
				queries := context.authorizer.GetNativeQueries()

				q, ok := queries[queryName]
				if !ok {
					return "--an error occurred during render--"
				}

				rawRes, err := context.storage.NativeQuery(q, args)
				if err != nil {
					return "--an error occurred during render--"
				}

				res, err := json.Marshal(rawRes)
				if err != nil {
					return "--an error occurred during render--"
				}

				return string(res)
			}
			return context.authorizer.Authorize(c, &authzCtx)
		} else {
			return c.JSON(&fiber.Map{"id": id})
		}
	}
}
