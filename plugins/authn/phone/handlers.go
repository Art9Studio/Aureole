package phone

import (
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
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

		specs := i.Collection.Spec
		exist, err := context.storage.IsIdentityExist(context.identity, []storageT.Filter{
			{Name: specs.FieldsMap["phone"].Name, Value: identityData.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !exist {
			return sendError(c, fiber.StatusUnauthorized, "user doesn't exist")
		}

		v := context.conf.Verification
		otp, err := getRandomString(v.Otp.Length, v.Otp.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otpHash, err := context.hasher.HashPw(v.Otp.Prefix + otp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verificationData := &storageT.PhoneVerificationData{
			Phone:    identityData.Phone,
			Otp:      otpHash,
			Attempts: 0,
			Expires:  time.Now().Add(time.Duration(v.Otp.Exp) * time.Second).Format(time.RFC3339),
			Invalid:  false,
		}

		vSpecs := &context.verification.coll.Spec
		err = context.storage.InvalidateVerification(vSpecs, []storageT.Filter{
			{Name: vSpecs.FieldsMap["phone"].Name, Value: identityData.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verificationId, err := context.storage.InsertVerification(vSpecs, verificationData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = context.verification.sender.Send(verificationData.Phone.(string),
			"",
			context.conf.Verification.Template,
			map[string]interface{}{"otp": otp})
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

		specs := context.identity.Collection.Spec
		exist, err := context.storage.IsIdentityExist(context.identity, []storageT.Filter{
			{Name: specs.FieldsMap["phone"].Name, Value: identityData.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if exist {
			return sendError(c, fiber.StatusBadRequest, "user already exist")
		}

		_, err = context.storage.InsertIdentity(context.identity, identityData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		v := context.conf.Verification
		otp, err := getRandomString(v.Otp.Length, v.Otp.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otpHash, err := context.hasher.HashPw(v.Otp.Prefix + otp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verificationData := &storageT.PhoneVerificationData{
			Phone:    identityData.Phone,
			Otp:      otpHash,
			Attempts: 0,
			Expires:  time.Now().Add(time.Duration(v.Otp.Exp) * time.Second).Format(time.RFC3339),
			Invalid:  false,
		}

		vSpecs := &context.verification.coll.Spec
		err = context.storage.InvalidateVerification(vSpecs, []storageT.Filter{
			{Name: vSpecs.FieldsMap["phone"].Name, Value: identityData.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verificationId, err := context.storage.InsertVerification(&context.verification.coll.Spec, verificationData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = context.verification.sender.Send(verificationData.Phone.(string),
			"",
			context.conf.Verification.Template,
			map[string]interface{}{"otp": otp})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"verification_id": verificationId})
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
		if statusCode, err := getJsonData(authInput, v.FieldsMap["otp"], &requestData.Otp); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		storage := context.storage
		vSpecs := context.verification.coll.Spec

		rawVerification, err := storage.GetVerification(&vSpecs, []storageT.Filter{
			{Name: vSpecs.FieldsMap["id"].Name, Value: requestData.Id},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verification, ok := rawVerification.(map[string]interface{})
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "cannot get verification data from database")
		}

		if verification[vSpecs.FieldsMap["invalid"].Name].(bool) {
			return sendError(c, fiber.StatusUnauthorized, "invalid verification")
		}

		expires, err := time.Parse(time.RFC3339, verification[vSpecs.FieldsMap["expires"].Name].(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if time.Now().After(expires) {
			return sendError(c, fiber.StatusUnauthorized, "verification otp expire")
		}

		// todo: fix type conversion
		if int(verification[vSpecs.FieldsMap["attempts"].Name].(float64)) >= context.conf.Verification.MaxAttempts {
			return sendError(c, fiber.StatusUnauthorized, "too much attempts")
		}

		otp := v.Otp.Prefix + requestData.Otp.(string)
		isMatch, err := context.hasher.ComparePw(otp, verification[vSpecs.FieldsMap["otp"].Name].(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if isMatch {
			iCollSpec := context.identity.Collection.Spec

			rawIdentity, err := context.storage.GetIdentity(context.identity, []storageT.Filter{
				{Name: iCollSpec.FieldsMap["phone"].Name, Value: verification[vSpecs.FieldsMap["phone"].Name]},
			})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			i, ok := rawIdentity.(map[string]interface{})
			if !ok {
				return sendError(c, fiber.StatusInternalServerError, "cannot get identity from database")
			}

			err = context.storage.SetPhoneVerified(&iCollSpec, []storageT.Filter{
				{Name: iCollSpec.FieldsMap["phone"].Name, Value: i[iCollSpec.FieldsMap["phone"].Name]},
			})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
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
			if err := context.storage.IncrAttempts(&vSpecs, []storageT.Filter{
				{Name: vSpecs.FieldsMap["id"].Name, Value: requestData.Id},
			}); err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
			return sendError(c, fiber.StatusUnauthorized, "wrong otp")
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

		rawVerification, err := storage.GetVerification(&collSpec, []storageT.Filter{
			{Name: collSpec.FieldsMap["id"].Name, Value: requestData.Id},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		oldVerification, ok := rawVerification.(map[string]interface{})
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "cannot get verification data from database")
		}

		otp, err := getRandomString(v.Otp.Length, v.Otp.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otpHash, err := context.hasher.HashPw(v.Otp.Prefix + otp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verification := &storageT.PhoneVerificationData{
			Phone:    oldVerification[collSpec.FieldsMap["phone"].Name],
			Otp:      otpHash,
			Attempts: 0,
			Expires:  time.Now().Add(time.Duration(v.Otp.Exp) * time.Second).Format(time.RFC3339),
			Invalid:  false,
		}

		vSpecs := &context.verification.coll.Spec
		err = context.storage.InvalidateVerification(vSpecs, []storageT.Filter{
			{Name: vSpecs.FieldsMap["phone"].Name, Value: verification.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verificationId, err := context.storage.InsertVerification(vSpecs, verification)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = context.verification.sender.Send(verification.Phone.(string),
			"",
			context.conf.Verification.Template,
			map[string]interface{}{"otp": otp})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"verification_id": verificationId})
	}
}
