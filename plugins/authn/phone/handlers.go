package phone

import (
	"aureole/internal/encrypt"
	"aureole/internal/identity"
	"aureole/internal/jwt"
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
		encOtp, err := encrypt.Encrypt(p.conf.Otp.Prefix + randStr + p.conf.Otp.Postfix)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		err = p.pluginApi.SaveToService(input.Phone, encOtp, p.conf.Otp.Exp)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := jwt.CreateJWT(
			map[string]interface{}{
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

		return c.JSON(fiber.Map{"token": token})
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
