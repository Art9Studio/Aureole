package sms

import (
	"aureole/internal/encrypt"
	"aureole/internal/jwt"
	"aureole/internal/router"
	"github.com/gofiber/fiber/v2"
)

func Resend(s *sms) func(*fiber.Ctx) error {
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

		otp, err := encrypt.GetRandomString(s.conf.Otp.Length, s.conf.Otp.Alphabet)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := jwt.CreateJWT(
			map[string]interface{}{
				"phone":    phone,
				"attempts": 0,
			},
			s.conf.Otp.Exp)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		encOtp, err := encrypt.Encrypt(otp)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		_ = encOtp
		err = s.pluginApi.SaveToService(phone.(string), encOtp, s.conf.Otp.Exp)
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = s.sender.Send(phone.(string), "", s.conf.Template, map[string]interface{}{"otp": otp})
		if err != nil {
			return router.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(&fiber.Map{"token": token})
	}
}
