package email

import (
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
)

func GetMagicLink(e *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		i := e.identity
		loginMap := e.conf.Login.FieldsMap
		if !i.Email.IsEnabled || !isCredential(&i.Email) {
			return sendError(c, fiber.StatusInternalServerError, "expects 1 credential, 0 got")
		}

		identity := &storageT.IdentityData{}
		if statusCode, err := getJsonData(authInput, loginMap["email"], &identity.Email); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		emailCol := e.coll.Spec.FieldsMap["email"].Name
		exist, err := e.storage.IsIdentityExist(i, []storageT.Filter{
			{Name: emailCol, Value: identity.Email},
		})
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

		tokenHash := e.link.hasher().Sum(token.Bytes())
		linkData := &storageT.EmailLinkData{
			Email:   identity.Email,
			Token:   base64.StdEncoding.EncodeToString(tokenHash),
			Expires: time.Now().Add(time.Duration(e.conf.Link.Token.Exp) * time.Second).Format(time.RFC3339),
			Invalid: false,
		}

		linkSpecs := &e.link.coll.Spec
		err = e.storage.InvalidateEmailLink(linkSpecs, []storageT.Filter{
			{Name: linkSpecs.FieldsMap["email"].Name, Value: identity.Email},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = e.storage.InsertEmailLink(linkSpecs, linkData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := initMagicLink(e.link.magicLink, token.String())
		err = e.link.sender.Send(linkData.Email.(string),
			"",
			e.conf.Link.Template,
			map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func Register(e *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		identity := &storageT.IdentityData{Additional: map[string]interface{}{}}
		if statusCode, err := getRegisterData(e, authInput, e.conf.Register.FieldsMap, identity); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		i := e.identity
		emailField := e.coll.Spec.FieldsMap["email"].Name
		exist, err := e.storage.IsIdentityExist(i, []storageT.Filter{
			{Name: emailField, Value: identity.Email},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if exist {
			return sendError(c, fiber.StatusBadRequest, "user already exist")
		}

		userId, err := e.storage.InsertIdentity(i, identity)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := uuid.NewV4()
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		tokenHash := e.link.hasher().Sum(token.Bytes())
		linkData := &storageT.EmailLinkData{
			Email:   identity.Email,
			Token:   base64.StdEncoding.EncodeToString(tokenHash),
			Expires: time.Now().Add(time.Duration(e.conf.Link.Token.Exp) * time.Second).Format(time.RFC3339),
			Invalid: false,
		}

		linkSpecs := &e.link.coll.Spec
		err = e.storage.InvalidateEmailLink(linkSpecs, []storageT.Filter{
			{Name: linkSpecs.FieldsMap["email"].Name, Value: identity.Email},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = e.storage.InsertEmailLink(linkSpecs, linkData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := initMagicLink(e.link.magicLink, token.String())
		err = e.link.sender.Send(linkData.Email.(string),
			"",
			e.conf.Link.Template,
			map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"user_id": userId})
	}
}

func Login(e *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		t := c.Query("token")
		if t == "" {
			return sendError(c, fiber.StatusNotFound, "token not found")
		}

		token, err := uuid.FromString(strings.TrimRight(t, "\n"))
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		tokenHash := e.link.hasher().Sum(token.Bytes())

		linkSpecs := &e.link.coll.Spec
		tokenName := e.link.coll.Spec.FieldsMap["token"].Name
		rawEmailLink, err := e.storage.GetEmailLink(linkSpecs, []storageT.Filter{
			{Name: tokenName, Value: base64.StdEncoding.EncodeToString(tokenHash)},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, fmt.Sprintf("%s: %s", err.Error(), token.String()))
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

		err = e.storage.InvalidateEmailLink(linkSpecs, []storageT.Filter{
			{Name: tokenName, Value: base64.StdEncoding.EncodeToString(tokenHash)},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		iCollSpec := e.identity.Collection.Spec
		rawIdentity, err := e.storage.GetIdentity(e.identity, []storageT.Filter{
			{Name: iCollSpec.FieldsMap["email"].Name, Value: emailLink[linkSpecs.FieldsMap["email"].Name]},
		})
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
			queries := e.authorizer.GetNativeQueries()

			q, ok := queries[queryName]
			if !ok {
				return "--an error occurred during render--"
			}

			rawRes, err := e.storage.NativeQuery(q, args...)
			if err != nil {
				return "--an error occurred during render--"
			}

			res, err := json.Marshal(rawRes)
			if err != nil {
				return "--an error occurred during render--"
			}

			return string(res)
		}
		return e.authorizer.Authorize(c, authzCtx)
	}
}
