package phone

import (
	"aureole/internal/encrypt"
	"aureole/internal/identity"
	"aureole/internal/jwt"
	authzT "aureole/internal/plugins/authz/types"
	"aureole/internal/router"
	"github.com/gofiber/fiber/v2"
)

func SendOtp(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Phone == "" {
			return router.SendError(c, fiber.StatusBadRequest, "phone required")
		}
		i := identity.Identity{Phone: input.Phone}

		randStr, err := encrypt.GetRandomString(p.conf.Otp.Length, p.conf.Otp.Alphabet)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otp := p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix
		otpHash, err := p.hasher.HashPw(otp)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := jwt.CreateJWT(
			map[string]interface{}{
				"otp":      otpHash,
				"phone":    i.Phone,
				"attempts": 0,
			},
			p.conf.Otp.Exp)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.sender.Send(i.Phone, "", p.conf.Template, map[string]interface{}{"otp": otp})
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"token": token})
	}
}

func Login(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Token == "" || input.Otp == "" {
			return router.SendError(c, fiber.StatusBadRequest, "token and otp are required")
		}

		t, err := jwt.ParseJWT(input.Token)
		if err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		phone, ok := t.Get("phone")
		if !ok {
			return router.SendError(c, fiber.StatusBadRequest, "cannot get phone from token")
		}
		otpHash, ok := t.Get("otp")
		if !ok {
			return router.SendError(c, fiber.StatusBadRequest, "cannot get otp from token")
		}
		attempts, ok := t.Get("attempts")
		if !ok {
			return router.SendError(c, fiber.StatusBadRequest, "cannot get attempts from token")
		}
		if err := jwt.InvalidateJWT(t); err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if int(attempts.(float64)) >= p.conf.MaxAttempts {
			return router.SendError(c, fiber.StatusUnauthorized, "too much attempts")
		}

		isMatch, err := p.hasher.ComparePw(input.Otp, otpHash.(string))
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if isMatch {
			var i = make(map[string]interface{})
			if p.manager != nil {
				i, err = p.manager.OnUserAuthenticated(
					&identity.Credential{
						Name:  identity.Phone,
						Value: phone.(string),
					},
					&identity.Identity{
						Phone:         phone.(string),
						PhoneVerified: true,
					},
					AdapterName,
					nil)
				if err != nil {
					return router.SendError(c, fiber.StatusInternalServerError, err.Error())
				}
			} else {
				i["phone"] = phone.(string)
			}

			return p.authorizer.Authorize(c, authzT.NewPayload(p.authorizer, nil, i))
		} else {
			token, err := jwt.CreateJWT(
				map[string]interface{}{
					"otp":      otpHash,
					"phone":    phone,
					"attempts": int(attempts.(float64)) + 1,
				},
				p.conf.Otp.Exp)
			if err != nil {
				return router.SendError(c, fiber.StatusInternalServerError, err.Error())
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"token": token})
		}
	}
}

func Resend(p *phone) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var input *input
		if err := c.BodyParser(input); err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Token == "" {
			return router.SendError(c, fiber.StatusBadRequest, "token are required")
		}

		t, err := jwt.ParseJWT(input.Token)
		if err != nil {
			return router.SendError(c, fiber.StatusBadRequest, err.Error())
		}
		phone, ok := t.Get("phone")
		if !ok {
			return router.SendError(c, fiber.StatusBadRequest, "cannot get phone from token")
		}
		if err := jwt.InvalidateJWT(t); err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		randStr, err := encrypt.GetRandomString(p.conf.Otp.Length, p.conf.Otp.Alphabet)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		otp := p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix
		otpHash, err := p.hasher.HashPw(otp)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := jwt.CreateJWT(
			map[string]interface{}{
				"otp":      otpHash,
				"phone":    phone,
				"attempts": 0,
			},
			p.conf.Otp.Exp)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.sender.Send(phone.(string), "", p.conf.Template, map[string]interface{}{"otp": otp})
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(&fiber.Map{"token": token})
	}
}
