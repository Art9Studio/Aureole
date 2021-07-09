package phone

import (
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"time"
)

func Login(context *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		i := context.identity
		loginMap := context.conf.Login.FieldsMap

		if !i.Phone.IsEnabled || !isCredential(&i.Phone) {
			return sendError(c, fiber.StatusInternalServerError, "expects 1 credential, 0 got")
		}

		identityData := &storageT.IdentityData{}
		if statusCode, err := getJsonData(authInput, loginMap["phone"], &identityData.Phone); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		exist, err := context.storage.IsIdentityExist(context.identity, "phone", identityData.Phone)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !exist {
			return sendError(c, fiber.StatusUnauthorized, "user doesn't exist")
		}

		v := context.conf.Verification
		code, err := getRandomString(v.Code.Length, v.Code.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		codeHash, err := context.hasher.HashPw(v.Code.Prefix + code)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verificationData := &storageT.PhoneVerificationData{
			Phone:    identityData.Phone,
			Code:     codeHash,
			Attempts: 0,
			Expires:  time.Now().Add(time.Duration(v.Code.Exp) * time.Second).Format(time.RFC3339),
		}
		verificationId, err := context.storage.InsertVerification(&context.verification.coll.Spec, verificationData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = context.verification.sender.Send(verificationData.Phone.(string),
			"",
			context.conf.Verification.Template,
			map[string]interface{}{"code": code})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"verification_id": verificationId})
	}
}

func Register(context *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		identityData := &storageT.IdentityData{Additional: map[string]interface{}{}}
		if statusCode, err := getRegisterData(context, authInput, context.conf.Register.FieldsMap, identityData); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		exist, err := context.storage.IsIdentityExist(context.identity, "phone", identityData.Phone)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if exist {
			return sendError(c, fiber.StatusBadRequest, "user already exist")
		}

		userId, err := context.storage.InsertIdentity(context.identity, identityData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if context.conf.Register.IsLoginAfter {
			v := context.conf.Verification
			code, err := getRandomString(v.Code.Length, v.Code.Alphabet)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			codeHash, err := context.hasher.HashPw(v.Code.Prefix + code)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			verificationData := &storageT.PhoneVerificationData{
				Phone:    identityData.Phone,
				Code:     codeHash,
				Attempts: 0,
				Expires:  time.Now().Add(time.Duration(v.Code.Exp) * time.Second).Format(time.RFC3339),
			}
			verificationId, err := context.storage.InsertVerification(&context.verification.coll.Spec, verificationData)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			err = context.verification.sender.Send(verificationData.Phone.(string),
				"",
				context.conf.Verification.Template,
				map[string]interface{}{"code": code})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			return c.JSON(&fiber.Map{"verification_id": verificationId})
		} else {
			return c.JSON(&fiber.Map{"id": userId})
		}
	}
}

func Verify(context *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		v := context.conf.Verification
		requestData := &storageT.PhoneVerificationData{}

		if statusCode, err := getJsonData(authInput, v.FieldsMap["id"], &requestData.Id); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		if statusCode, err := getJsonData(authInput, v.FieldsMap["code"], &requestData.Code); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		storage := context.storage
		collSpec := context.verification.coll.Spec

		rawVerificationData, err := storage.GetVerification(&collSpec, collSpec.FieldsMap["id"].Name, requestData.Id)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verificationData, ok := rawVerificationData.(map[string]interface{})
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "cannot get verification data from database")
		}

		expires, err := time.Parse(time.RFC3339, verificationData[collSpec.FieldsMap["expires"].Name].(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if time.Now().After(expires) {
			return sendError(c, fiber.StatusUnauthorized, "verification code expire")
		}

		// todo: fix type conversion
		if int(verificationData[collSpec.FieldsMap["attempts"].Name].(float64)) >= context.conf.Verification.MaxAttempts {
			return sendError(c, fiber.StatusUnauthorized, "too much attempts")
		}

		code := v.Code.Prefix + requestData.Code.(string)
		isMatch, err := context.hasher.ComparePw(code, verificationData[collSpec.FieldsMap["code"].Name].(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if isMatch {
			iCollSpec := context.identity.Collection.Spec

			rawIdentity, err := context.storage.GetIdentity(context.identity,
				iCollSpec.FieldsMap["phone"].Name,
				verificationData[collSpec.FieldsMap["phone"].Name],
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
		} else {
			if err := context.storage.IncrAttempts(&collSpec, collSpec.FieldsMap["id"].Name, requestData.Id); err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
			return sendError(c, fiber.StatusUnauthorized, "wrong verification code")
		}

	}
}

func Resend(context *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		v := context.conf.Verification
		requestData := &storageT.PhoneVerificationData{}

		statusCode, err := getJsonData(authInput, v.FieldsMap["id"], &requestData.Id)
		if err != nil {
			return sendError(c, statusCode, err.Error())
		}

		storage := context.storage
		collSpec := context.verification.coll.Spec

		rawVerificationData, err := storage.GetVerification(&collSpec, collSpec.FieldsMap["id"].Name, requestData.Id)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		oldVerificationData, ok := rawVerificationData.(map[string]interface{})
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "cannot get verification data from database")
		}

		code, err := getRandomString(v.Code.Length, v.Code.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		codeHash, err := context.hasher.HashPw(v.Code.Prefix + code)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verificationData := &storageT.PhoneVerificationData{
			Phone:    oldVerificationData[collSpec.FieldsMap["phone"].Name],
			Code:     codeHash,
			Attempts: 0,
			Expires:  time.Now().Add(time.Duration(v.Code.Exp) * time.Second).Format(time.RFC3339),
		}
		verificationId, err := context.storage.InsertVerification(&context.verification.coll.Spec, verificationData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = context.verification.sender.Send(verificationData.Phone.(string),
			"",
			context.conf.Verification.Template,
			map[string]interface{}{"code": code})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"verification_id": verificationId})
	}
}
