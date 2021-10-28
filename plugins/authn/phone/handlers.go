package phone

import (
	authnTypes "aureole/internal/plugins/authn/types"
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
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
		identityData := &storageT.IdentityData{Phone: input.Phone}

		/*exist, err := p.storage.IsIdentityExist(p.identity, []storageT.Filter{
			{Name: p.identity.Collection.Spec.FieldsMap["phone"].Name, Value: identityData.Phone},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !exist {
			return sendError(c, fiber.StatusUnauthorized, "user doesn't exist")
		}*/

		randStr, err := getRandomString(p.conf.Otp.Length, p.conf.Otp.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otp := p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix
		otpHash, err := p.hasher.HashPw(otp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := createToken(p, map[string]interface{}{
			"otp":      otpHash,
			"phone":    identityData.Phone,
			"attempts": 0,
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.sender.Send(identityData.Phone.(string),
			"",
			p.conf.Template,
			map[string]interface{}{"otp": otp})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"token": token})
	}
}

/*func Register(p *phone) func(*fiber.Ctx) error {
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
}*/

func Login(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		input, err := authnTypes.NewInput(c)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Additional["token"] == nil || input.Additional["otp"] == nil {
			return sendError(c, fiber.StatusBadRequest, "token and otp are required")
		}

		t, err := jwt.ParseString(
			input.Additional["token"].(string),
			jwt.WithIssuer("Aureole Internal"),
			jwt.WithAudience("Aureole Internal"),
			jwt.WithValidate(true),
			jwt.WithKeySet(p.serviceKey.GetPublicSet()),
		)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		phone, ok := t.Get("phone")
		if !ok {
			return sendError(c, fiber.StatusBadRequest, "cannot get phone from token")
		}
		otpHash, ok := t.Get("otp")
		if !ok {
			return sendError(c, fiber.StatusBadRequest, "cannot get otp from token")
		}
		attempts, ok := t.Get("attempts")
		if !ok {
			return sendError(c, fiber.StatusBadRequest, "cannot get attempts from token")
		}

		if int(attempts.(float64)) >= p.conf.MaxAttempts {
			return sendError(c, fiber.StatusUnauthorized, "too much attempts")
		}

		isMatch, err := p.hasher.ComparePw(input.Additional["otp"].(string), otpHash.(string))
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
			payload := authzT.NewPayload(p.authorizer, p.storage)
			payload.Phone = phone
			return p.authorizer.Authorize(c, payload)
		} else {
			token, err := createToken(p, map[string]interface{}{
				"otp":      otpHash,
				"phone":    phone,
				"attempts": int(attempts.(float64)) + 1,
			})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"token": token})
		}
	}
}

func Resend(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		input, err := authnTypes.NewInput(c)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Additional["token"] == nil {
			return sendError(c, fiber.StatusBadRequest, "token are required")
		}

		t, err := jwt.ParseString(
			input.Additional["token"].(string),
			jwt.WithIssuer("Aureole Internal"),
			jwt.WithAudience("Aureole Internal"),
			jwt.WithValidate(true),
			jwt.WithKeySet(p.serviceKey.GetPublicSet()),
		)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		phone, ok := t.Get("phone")
		if !ok {
			return sendError(c, fiber.StatusBadRequest, "cannot get phone from token")
		}

		randStr, err := getRandomString(p.conf.Otp.Length, p.conf.Otp.Alphabet)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otp := p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix
		otpHash, err := p.hasher.HashPw(otp)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := createToken(p, map[string]interface{}{
			"otp":      otpHash,
			"phone":    phone,
			"attempts": 0,
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.sender.Send(phone.(string),
			"",
			p.conf.Template,
			map[string]interface{}{"otp": otp})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"token": token})
	}
}
