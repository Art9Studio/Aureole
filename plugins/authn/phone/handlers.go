package phone

import (
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Login(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		i := p.identity
		loginMap := p.conf.Login.FieldsMap

		if !i.Phone.IsEnabled || !isCredential(&i.Phone) {
			return sendError(c, fiber.StatusInternalServerError, "expects 1 credential, 0 got")
		}

		identityData := &storageT.IdentityData{}
		if statusCode, err := getJsonData(authInput, loginMap["phone"], &identityData.Phone); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		specs := i.Collection.Spec
		exist, err := p.storage.IsIdentityExist(i, []storageT.Filter{
			{Name: specs.FieldsMap["phone"].Name, Value: identityData.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !exist {
			return sendError(c, fiber.StatusUnauthorized, "user doesn't exist")
		}

		v := p.conf.Verification
		otp, err := getRandomString(v.Otp.Length, v.Otp.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otpHash, err := p.hasher.HashPw(v.Otp.Prefix + otp)
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

		vSpecs := &p.verification.coll.Spec
		err = p.storage.InvalidateVerification(vSpecs, []storageT.Filter{
			{Name: vSpecs.FieldsMap["phone"].Name, Value: identityData.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verificationId, err := p.storage.InsertVerification(vSpecs, verificationData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.verification.sender.Send(verificationData.Phone.(string),
			"",
			p.conf.Verification.Template,
			map[string]interface{}{"otp": otp})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"verification_id": verificationId})
	}
}

func Register(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		identityData := &storageT.IdentityData{Additional: map[string]interface{}{}}
		if statusCode, err := getRegisterData(p, authInput, p.conf.Register.FieldsMap, identityData); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		i := p.identity
		specs := i.Collection.Spec
		exist, err := p.storage.IsIdentityExist(i, []storageT.Filter{
			{Name: specs.FieldsMap["phone"].Name, Value: identityData.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if exist {
			return sendError(c, fiber.StatusBadRequest, "user already exist")
		}

		_, err = p.storage.InsertIdentity(i, identityData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		v := p.conf.Verification
		otp, err := getRandomString(v.Otp.Length, v.Otp.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otpHash, err := p.hasher.HashPw(v.Otp.Prefix + otp)
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

		vSpecs := &p.verification.coll.Spec
		err = p.storage.InvalidateVerification(vSpecs, []storageT.Filter{
			{Name: vSpecs.FieldsMap["phone"].Name, Value: identityData.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verificationId, err := p.storage.InsertVerification(&p.verification.coll.Spec, verificationData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.verification.sender.Send(verificationData.Phone.(string),
			"",
			p.conf.Verification.Template,
			map[string]interface{}{"otp": otp})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"verification_id": verificationId})
	}
}

func Verify(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		v := p.conf.Verification
		requestData := &storageT.PhoneVerificationData{}
		if statusCode, err := getJsonData(authInput, v.FieldsMap["id"], &requestData.Id); err != nil {
			return sendError(c, statusCode, err.Error())
		}
		if statusCode, err := getJsonData(authInput, v.FieldsMap["otp"], &requestData.Otp); err != nil {
			return sendError(c, statusCode, err.Error())
		}

		storage := p.storage
		vSpecs := p.verification.coll.Spec

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
		if int(verification[vSpecs.FieldsMap["attempts"].Name].(float64)) >= p.conf.Verification.MaxAttempts {
			return sendError(c, fiber.StatusUnauthorized, "too much attempts")
		}

		otp := v.Otp.Prefix + requestData.Otp.(string)
		isMatch, err := p.hasher.ComparePw(otp, verification[vSpecs.FieldsMap["otp"].Name].(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if isMatch {
			i := p.identity
			iCollSpec := i.Collection.Spec

			rawIdentity, err := p.storage.GetIdentity(i, []storageT.Filter{
				{Name: iCollSpec.FieldsMap["phone"].Name, Value: verification[vSpecs.FieldsMap["phone"].Name]},
			})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			identity, ok := rawIdentity.(map[string]interface{})
			if !ok {
				return sendError(c, fiber.StatusInternalServerError, "cannot get identity from database")
			}

			err = p.storage.SetPhoneVerified(&iCollSpec, []storageT.Filter{
				{Name: iCollSpec.FieldsMap["phone"].Name, Value: identity[iCollSpec.FieldsMap["phone"].Name]},
			})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			authzCtx := authzT.NewContext(identity, iCollSpec.FieldsMap)
			// todo: refactor this
			authzCtx.NativeQ = func(queryName string, args ...interface{}) string {
				queries := p.authorizer.GetNativeQueries()

				q, ok := queries[queryName]
				if !ok {
					return "--an error occurred during render--"
				}

				rawRes, err := p.storage.NativeQuery(q, args...)
				if err != nil {
					return "--an error occurred during render--"
				}

				res, err := json.Marshal(rawRes)
				if err != nil {
					return "--an error occurred during render--"
				}

				return string(res)
			}
			return p.authorizer.Authorize(c, authzCtx)
		} else {
			if err := p.storage.IncrAttempts(&vSpecs, []storageT.Filter{
				{Name: vSpecs.FieldsMap["id"].Name, Value: requestData.Id},
			}); err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
			return sendError(c, fiber.StatusUnauthorized, "wrong otp")
		}
	}
}

func Resend(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var authInput interface{}
		if err := c.BodyParser(&authInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		v := p.conf.Verification
		requestData := &storageT.PhoneVerificationData{}

		statusCode, err := getJsonData(authInput, v.FieldsMap["id"], &requestData.Id)
		if err != nil {
			return sendError(c, statusCode, err.Error())
		}

		storage := p.storage
		collSpec := p.verification.coll.Spec

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

		otpHash, err := p.hasher.HashPw(v.Otp.Prefix + otp)
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

		vSpecs := &p.verification.coll.Spec
		err = p.storage.InvalidateVerification(vSpecs, []storageT.Filter{
			{Name: vSpecs.FieldsMap["phone"].Name, Value: verification.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verificationId, err := p.storage.InsertVerification(vSpecs, verification)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.verification.sender.Send(verification.Phone.(string),
			"",
			p.conf.Verification.Template,
			map[string]interface{}{"otp": otp})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"verification_id": verificationId})
	}
}
