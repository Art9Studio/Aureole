package phone

import (
	authnTypes "aureole/internal/plugins/authn/types"
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SendOtp(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if !p.identity.Phone.IsEnabled || !isCredential(&p.identity.Phone) {
			return sendError(c, fiber.StatusInternalServerError, "expects 1 credential, 0 got")
		}

		input, err := authnTypes.NewInput(c)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Phone == nil {
			return sendError(c, fiber.StatusBadRequest, "phone required")
		}
		identityData := &storageT.IdentityData{Email: input.Phone}

		/*exist, err := p.storage.IsIdentityExist(p.identity, []storageT.Filter{
			{Name: p.identity.Collection.Spec.FieldsMap["phone"].Name, Value: identityData.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !exist {
			return sendError(c, fiber.StatusUnauthorized, "user doesn't exist")
		}*/

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
		input, err := authnTypes.NewInput(c)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if err := input.Init(p.identity, p.identity.Collection.Spec.FieldsMap, true); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		identityData := &storageT.IdentityData{
			Id:         input.Id,
			Username:   input.Username,
			Phone:      input.Phone,
			Email:      input.Email,
			Additional: input.Additional,
		}

		specs := p.identity.Collection.Spec
		exist, err := p.storage.IsIdentityExist(p.identity, []storageT.Filter{
			{Name: specs.FieldsMap["phone"].Name, Value: identityData.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if exist {
			return sendError(c, fiber.StatusBadRequest, "user already exist")
		}

		_, err = p.storage.InsertIdentity(p.identity, identityData)
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

func Login(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		input, err := authnTypes.NewInput(c)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Additional["id"] == nil || input.Additional["otp"] == nil {
			return sendError(c, fiber.StatusBadRequest, "id and otp are required")
		}
		requestData := storageT.PhoneVerificationData{
			Id:  input.Additional["id"],
			Otp: input.Additional["otp"],
		}

		vSpecs := p.verification.coll.Spec
		rawVerification, err := p.storage.GetVerification(&vSpecs, []storageT.Filter{
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

		otp := p.conf.Verification.Otp.Prefix + requestData.Otp.(string)
		isMatch, err := p.hasher.ComparePw(otp, verification[vSpecs.FieldsMap["otp"].Name].(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if isMatch {
			/*i := p.identity
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

			payload := authzT.NewPayload(identity, iCollSpec.FieldsMap)*/
			// todo: refactor this
			payload := &authzT.Payload{Phone: verification[vSpecs.FieldsMap["phone"].Name]}
			payload.NativeQ = func(queryName string, args ...interface{}) string {
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
			return p.authorizer.Authorize(c, payload)
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
		input, err := authnTypes.NewInput(c)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Additional["id"] == nil {
			return sendError(c, fiber.StatusBadRequest, "id is required")
		}
		requestData := storageT.PhoneVerificationData{Id: input.Additional["id"]}

		collSpec := p.verification.coll.Spec
		rawVerification, err := p.storage.GetVerification(&collSpec, []storageT.Filter{
			{Name: collSpec.FieldsMap["id"].Name, Value: requestData.Id},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		oldVerification, ok := rawVerification.(map[string]interface{})
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "cannot get verification data from database")
		}

		otp, err := getRandomString(p.conf.Verification.Otp.Length, p.conf.Verification.Otp.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otpHash, err := p.hasher.HashPw(p.conf.Verification.Otp.Prefix + otp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		verification := &storageT.PhoneVerificationData{
			Phone:    oldVerification[collSpec.FieldsMap["phone"].Name],
			Otp:      otpHash,
			Attempts: 0,
			Expires:  time.Now().Add(time.Duration(p.conf.Verification.Otp.Exp) * time.Second).Format(time.RFC3339),
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
