package email

import (
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"encoding/base64"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"time"
)

func GetMagicLink(context *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		i := context.identity
		loginMap := context.conf.Login.FieldsMap
		if !i.Email.IsEnabled || !isCredential(&i.Email) {
			return sendError(c, fiber.StatusInternalServerError, "expects 1 credential, 0 got")
		}

		identity := &storageT.IdentityData{}
		if statusCode, err := getJsonData(authInput, loginMap["email"], &identity.Email); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		emailCol := context.coll.Spec.FieldsMap["email"].Name
		exist, err := context.storage.IsIdentityExist(context.identity, emailCol, identity.Email)
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

		tokenHash := context.link.hasher().Sum([]byte(token.String()))
		linkData := &storageT.EmailLinkData{
			Email:   identity.Email,
			Token:   base64.StdEncoding.EncodeToString(tokenHash),
			Expires: time.Now().Add(time.Duration(context.conf.Link.Token.Exp) * time.Second).Format(time.RFC3339),
			Invalid: false,
		}

		linkSpecs := &context.link.coll.Spec
		err = context.storage.InvalidateEmailLink(linkSpecs, linkSpecs.FieldsMap["email"].Name, identity.Phone)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = context.storage.InsertEmailLink(linkSpecs, linkData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := getMagicLink(context, token.String())
		err = context.link.sender.Send(linkData.Email.(string),
			"",
			context.conf.Link.Template,
			map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func Register(context *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		identity := &storageT.IdentityData{Additional: map[string]interface{}{}}
		if statusCode, err := getRegisterData(context, authInput, context.conf.Register.FieldsMap, identity); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		emailField := context.coll.Spec.FieldsMap["email"].Name
		exist, err := context.storage.IsIdentityExist(context.identity, emailField, identity.Email)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if exist {
			return sendError(c, fiber.StatusBadRequest, "user already exist")
		}

		userId, err := context.storage.InsertIdentity(context.identity, identity)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := uuid.NewV4()
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		tokenHash := context.link.hasher().Sum([]byte(token.String()))
		linkData := &storageT.EmailLinkData{
			Email:   identity.Email,
			Token:   base64.StdEncoding.EncodeToString(tokenHash),
			Expires: time.Now().Add(time.Duration(context.conf.Link.Token.Exp) * time.Second).Format(time.RFC3339),
			Invalid: false,
		}

		linkSpecs := &context.link.coll.Spec
		err = context.storage.InvalidateEmailLink(linkSpecs, linkSpecs.FieldsMap["email"].Name, identity.Email)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = context.storage.InsertEmailLink(linkSpecs, linkData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := getMagicLink(context, token.String())
		err = context.link.sender.Send(linkData.Email.(string),
			"",
			context.conf.Link.Template,
			map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"user_id": userId})
	}
}

func Login(context *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		token := c.Query("token")
		if token == "" {
			return sendError(c, fiber.StatusNotFound, "page not found")
		}

		linkSpecs := &context.link.coll.Spec
		tokenName := context.link.coll.Spec.FieldsMap["token"].Name

		tokenHash := context.link.hasher().Sum([]byte(token))
		rawEmailLink, err := context.storage.GetEmailLink(linkSpecs, tokenName, base64.StdEncoding.EncodeToString(tokenHash))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		emailLink, ok := rawEmailLink.(map[string]interface{})
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "cannot get magic link data from database")
		}

		if emailLink[linkSpecs.FieldsMap["invalid"].Name].(bool) {
			return sendError(c, fiber.StatusUnauthorized, "invalid token")
		}

		expires, err := time.Parse(time.RFC3339, emailLink[linkSpecs.FieldsMap["expires"].Name].(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if time.Now().After(expires) {
			return sendError(c, fiber.StatusUnauthorized, "link expire")
		}

		err = context.storage.InvalidateEmailLink(linkSpecs, tokenName, base64.StdEncoding.EncodeToString(tokenHash))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		iCollSpec := context.identity.Collection.Spec
		rawIdentity, err := context.storage.GetIdentity(context.identity,
			iCollSpec.FieldsMap["email"].Name,
			emailLink[linkSpecs.FieldsMap["email"].Name],
		)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		i, ok := rawIdentity.(map[string]interface{})
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "cannot get identity from database")
		}

		authzCtx := authzT.NewContext(i, iCollSpec.FieldsMap)
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
	}
}
