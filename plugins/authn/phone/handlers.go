package phone

import (
	"aureole/internal/identity"
	authzT "aureole/internal/plugins/authz/types"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
)

func SendOtp(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Phone == "" {
			return sendError(c, fiber.StatusBadRequest, "phone required")
		}
		i := identity.Identity{Phone: input.Phone}

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
			"phone":    i.Phone,
			"attempts": 0,
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.sender.Send(i.Phone, "", p.conf.Template, map[string]interface{}{"otp": otp})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"token": token})
	}
}

func Login(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Token == "" || input.Otp == "" {
			return sendError(c, fiber.StatusBadRequest, "token and otp are required")
		}

		t, err := jwt.ParseString(
			input.Token,
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

		isMatch, err := p.hasher.ComparePw(input.Otp, otpHash.(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if isMatch {
			var i = make(map[string]interface{})
			if p.manager != nil {
				i, err = p.manager.OnUserAuthenticated(
					&identity.Credential{
						Name:  "phone",
						Value: phone,
					},
					&identity.Identity{
						Phone:         phone.(string),
						PhoneVerified: true,
					},
					AdapterName,
					nil)
				if err != nil {
					return sendError(c, fiber.StatusInternalServerError, err.Error())
				}
			} else {
				i["phone"] = phone.(string)
			}

			return p.authorizer.Authorize(c, authzT.NewPayload(p.authorizer, nil, i))
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
		var input *input
		if err := c.BodyParser(input); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Token == "" {
			return sendError(c, fiber.StatusBadRequest, "token are required")
		}

		t, err := jwt.ParseString(
			input.Token,
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

		err = p.sender.Send(phone.(string), "", p.conf.Template, map[string]interface{}{"otp": otp})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(&fiber.Map{"token": token})
	}
}
