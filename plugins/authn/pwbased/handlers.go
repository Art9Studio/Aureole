package pwbased

import (
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"time"
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

func Reset(context *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		identityData := &storageT.IdentityData{}
		collMap := context.coll.Parent.Spec.FieldsMap
		i := context.identity
		getLoginTraitData(&i.Email, authInput, context.conf.Login.FieldsMap["email"], collMap["email"].Default, &identityData.Email)

		exist, err := context.storage.IsIdentityExist(context.identity, collMap["email"].Name, identityData.Email)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if !exist {
			return sendError(c, fiber.StatusUnauthorized, "user doesn't exist")
		}

		token, err := uuid.NewV4()
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		tokenHash := context.reset.hasher().Sum([]byte(token.String()))
		resetData := &storageT.PwResetData{
			Email:   identityData.Email,
			Token:   base64.StdEncoding.EncodeToString(tokenHash),
			Expires: time.Now().Add(time.Duration(context.conf.Reset.Token.Exp) * time.Second).Format(time.RFC3339),
			Invalid: false,
		}

		collSpec := &context.reset.coll.Spec
		err = context.storage.InvalidateReset(collSpec, collSpec.FieldsMap["email"].Name, identityData.Email)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = context.storage.InsertReset(&context.reset.coll.Spec, resetData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := fmt.Sprintf("http://localhost:3000/two/password/reset/confirm?token=%s", token.String())
		err = context.reset.sender.Send(resetData.Email.(string),
			"Reset your password",
			context.conf.Reset.Template,
			map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func ResetConfirm(context *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		token := c.Query("token")
		if token == "" {
			return sendError(c, fiber.StatusNotFound, "page not found")
		}

		resetSpecs := &context.reset.coll.Spec
		tokenName := context.reset.coll.Spec.FieldsMap["token"].Name

		tokenHash := context.reset.hasher().Sum([]byte(token))
		rawReset, err := context.storage.GetReset(resetSpecs, tokenName, base64.StdEncoding.EncodeToString(tokenHash))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		reset, ok := rawReset.(map[string]interface{})
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "cannot get reset data from database")
		}

		if reset[resetSpecs.FieldsMap["invalid"].Name].(bool) {
			return sendError(c, fiber.StatusUnauthorized, "invalid token")
		}

		expires, err := time.Parse(time.RFC3339, reset[resetSpecs.FieldsMap["expires"].Name].(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if time.Now().After(expires) {
			return sendError(c, fiber.StatusUnauthorized, "link expire")
		}

		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		r := context.conf.Reset
		pw := &storageT.PwBasedData{}

		if statusCode, err := getJsonData(authInput, r.FieldsMap["password"], &pw.Password); err != nil {
			return sendError(c, statusCode, err.Error())
		}
		pwHash, err := context.pwHasher.HashPw(pw.Password.(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		pw.PasswordHash = pwHash

		identitySpecs := context.coll.Parent.Spec
		email := reset[resetSpecs.FieldsMap["email"].Name].(string)
		_, err = context.storage.UpdatePassword(context.coll, identitySpecs.FieldsMap["email"].Name, email, pw.PasswordHash)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = context.storage.InvalidateReset(resetSpecs, tokenName, base64.StdEncoding.EncodeToString(tokenHash))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = context.reset.sender.SendRaw(reset[resetSpecs.FieldsMap["email"].Name].(string),
			"Reset your password",
			"Your password has been successfully changed")
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		// todo: add expiring any current user session

		return c.JSON(&fiber.Map{"status": "success"})
	}
}
